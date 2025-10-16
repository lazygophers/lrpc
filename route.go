package lrpc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Route struct {
	Method string
	Path   string

	Handler HandlerFunc

	Before, After []HandlerFunc

	// 可以存储一些类似于权限等信息，会在调用前写入到 local 中
	Extra map[string]any
}

// RouteType defines the type of route segment
type RouteType int

const (
	RouteTypeStatic   RouteType = iota // /users/list
	RouteTypeTyped                     // /users/{id:int}
	RouteTypeParam                     // /users/{id}
	RouteTypeWildcard                  // /users/*
	RouteTypeCatchAll                  // /users/**
)

// ParamConstraint defines parameter validation rules
type ParamConstraint struct {
	Type  string           // int, string, uuid, regex, digit, etc.
	Min   *int             // min value for int, min length for string
	Max   *int             // max value for int, max length for string
	Len   *int             // fixed length for string
	Regex *regexp.Regexp   // regex pattern
	Enum  []string         // enum values
}

// RouteParam defines a route parameter
type RouteParam struct {
	Name       string
	Constraint *ParamConstraint
}

// RouteNode represents a node in the route tree
type RouteNode struct {
	// Node properties
	typ      RouteType
	segment  string // static segment or full pattern
	param    *RouteParam
	priority int // for sorting: static > typed > param > wildcard > catchAll

	// Children nodes
	children []*RouteNode

	// Handler chain (middleware + handler)
	handlers []HandlerFunc

	// Full route pattern (for debugging)
	pattern string
}

// Router manages routes and middleware
type Router struct {
	root              *RouteNode
	middleware        []HandlerFunc // global middleware
	notFound          HandlerFunc
	methodNotAllowed  HandlerFunc
}

// MatchResult contains the result of a route match
type MatchResult struct {
	Node   *RouteNode
	Params map[string]string
}

type RouteOption func(r *Route)

func RouteWithPrefix(prefix string) RouteOption {
	return func(r *Route) {
		r.Path = prefix + r.Path
	}
}

func RouteWithBefore(before HandlerFunc) RouteOption {
	return func(r *Route) {
		r.Before = append(r.Before, before)
	}
}

func RouteWithAfter(after HandlerFunc) RouteOption {
	return func(r *Route) {
		r.After = append(r.After, after)
	}
}

func RouteWithExtra(extra map[string]any) RouteOption {
	return func(r *Route) {
		r.Extra = extra
	}
}

func RouteWithMergeExtra(merge map[string]any) RouteOption {
	return func(r *Route) {
		if r.Extra == nil {
			r.Extra = make(map[string]any)
		}

		for k, v := range merge {
			r.Extra[k] = v
		}
	}
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		root: &RouteNode{
			typ:      RouteTypeStatic,
			segment:  "/",
			children: make([]*RouteNode, 0),
		},
		middleware: make([]HandlerFunc, 0),
	}
}

// parseRoute parses a route pattern into nodes
func parseRoute(pattern string) ([]*RouteNode, error) {
	if pattern == "" || pattern[0] != '/' {
		return nil, fmt.Errorf("route must start with /")
	}

	// Handle root route
	if pattern == "/" {
		return []*RouteNode{{
			typ:     RouteTypeStatic,
			segment: "/",
			pattern: "/",
		}}, nil
	}

	// Split by /
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	nodes := make([]*RouteNode, 0, len(parts))

	for i, part := range parts {
		if part == "" {
			continue
		}

		node, err := parseSegment(part, i == len(parts)-1)
		if err != nil {
			return nil, fmt.Errorf("invalid segment %q: %w", part, err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseSegment parses a single segment
func parseSegment(segment string, isLast bool) (*RouteNode, error) {
	// Handle catchAll: **
	if segment == "**" {
		if !isLast {
			return nil, fmt.Errorf("** must be the last segment")
		}
		return &RouteNode{
			typ:      RouteTypeCatchAll,
			segment:  segment,
			priority: 4,
		}, nil
	}

	// Handle wildcard: *
	if segment == "*" {
		if !isLast {
			return nil, fmt.Errorf("* must be the last segment")
		}
		return &RouteNode{
			typ:      RouteTypeWildcard,
			segment:  segment,
			priority: 3,
		}, nil
	}

	// Handle parameter: {id} or {id:type} or :id
	if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
		return parseParamSegment(segment)
	}

	if strings.HasPrefix(segment, ":") {
		// Convert :id to {id}
		return parseParamSegment("{" + segment[1:] + "}")
	}

	// Check for mixed params like {uid}-{cid}
	if strings.Contains(segment, "{") {
		return parseMixedSegment(segment)
	}

	// Static segment
	return &RouteNode{
		typ:      RouteTypeStatic,
		segment:  segment,
		priority: 0,
	}, nil
}

// parseParamSegment parses {id} or {id:type,constraints}
func parseParamSegment(segment string) (*RouteNode, error) {
	// Remove { }
	inner := segment[1 : len(segment)-1]

	// Split by :
	parts := strings.SplitN(inner, ":", 2)
	paramName := parts[0]

	if paramName == "" {
		return nil, fmt.Errorf("parameter name cannot be empty")
	}

	node := &RouteNode{
		segment: segment,
		param: &RouteParam{
			Name: paramName,
		},
	}

	// No type constraint
	if len(parts) == 1 {
		node.typ = RouteTypeParam
		node.priority = 2
		return node, nil
	}

	// Parse type and constraints
	constraint, err := parseConstraint(parts[1])
	if err != nil {
		return nil, err
	}

	node.param.Constraint = constraint
	node.typ = RouteTypeTyped
	node.priority = 1

	return node, nil
}

// parseConstraint parses type and constraints like "int,min=5,max=10"
func parseConstraint(spec string) (*ParamConstraint, error) {
	parts := strings.Split(spec, ",")
	constraint := &ParamConstraint{
		Type: parts[0],
	}

	// Parse additional constraints
	for i := 1; i < len(parts); i++ {
		kv := strings.SplitN(parts[i], "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid constraint format: %s", parts[i])
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "min":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid min value: %s", value)
			}
			constraint.Min = &v

		case "max":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid max value: %s", value)
			}
			constraint.Max = &v

		case "len":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid len value: %s", value)
			}
			constraint.Len = &v

		case "enum":
			constraint.Enum = strings.Split(value, "|")

		default:
			return nil, fmt.Errorf("unknown constraint: %s", key)
		}
	}

	// Handle regex type
	if strings.HasPrefix(constraint.Type, "^") || strings.HasSuffix(constraint.Type, "$") {
		re, err := regexp.Compile(constraint.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		constraint.Regex = re
		constraint.Type = "regex"
	}

	return constraint, nil
}

// parseMixedSegment parses segments like {uid}-{cid}
func parseMixedSegment(segment string) (*RouteNode, error) {
	// This is a static+param mixed node
	// For simplicity, treat it as a static node with regex matching
	// Extract all {param} patterns
	pattern := segment
	params := make([]*RouteParam, 0)

	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(segment, -1)

	for _, match := range matches {
		paramDef := match[1]
		parts := strings.SplitN(paramDef, ":", 2)

		param := &RouteParam{
			Name: parts[0],
		}

		if len(parts) == 2 {
			constraint, err := parseConstraint(parts[1])
			if err != nil {
				return nil, err
			}
			param.Constraint = constraint
		}

		params = append(params, param)

		// Replace {param} with regex group
		if param.Constraint != nil && param.Constraint.Regex != nil {
			pattern = strings.Replace(pattern, match[0], "("+param.Constraint.Regex.String()+")", 1)
		} else {
			pattern = strings.Replace(pattern, match[0], "([^/]+)", 1)
		}
	}

	// Compile the final regex
	finalRegex, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil, fmt.Errorf("failed to compile mixed pattern: %w", err)
	}

	// Store the first param for now (TODO: support multiple params)
	node := &RouteNode{
		typ:      RouteTypeTyped,
		segment:  segment,
		priority: 1,
	}

	if len(params) > 0 {
		node.param = params[0]
		if node.param.Constraint == nil {
			node.param.Constraint = &ParamConstraint{}
		}
		node.param.Constraint.Regex = finalRegex
	}

	return node, nil
}

// validateParam validates a parameter value against its constraint
func validateParam(value string, constraint *ParamConstraint) error {
	if constraint == nil {
		return nil
	}

	switch constraint.Type {
	case "int", "digit":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parameter must be an integer")
		}

		if constraint.Min != nil && v < int64(*constraint.Min) {
			return fmt.Errorf("parameter must be >= %d", *constraint.Min)
		}

		if constraint.Max != nil && v > int64(*constraint.Max) {
			return fmt.Errorf("parameter must be <= %d", *constraint.Max)
		}

	case "number":
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("parameter must be a number")
		}

	case "string", "str":
		if constraint.Len != nil && len(value) != *constraint.Len {
			return fmt.Errorf("parameter length must be %d", *constraint.Len)
		}

		if constraint.Min != nil && len(value) < *constraint.Min {
			return fmt.Errorf("parameter length must be >= %d", *constraint.Min)
		}

		if constraint.Max != nil && len(value) > *constraint.Max {
			return fmt.Errorf("parameter length must be <= %d", *constraint.Max)
		}

	case "uuid":
		if _, err := uuid.Parse(value); err != nil {
			return fmt.Errorf("parameter must be a valid UUID")
		}

	case "bool":
		if value != "true" && value != "false" && value != "1" && value != "0" {
			return fmt.Errorf("parameter must be a boolean")
		}

	case "regex":
		if constraint.Regex != nil && !constraint.Regex.MatchString(value) {
			return fmt.Errorf("parameter does not match pattern")
		}
	}

	// Check enum
	if len(constraint.Enum) > 0 {
		found := false
		for _, v := range constraint.Enum {
			if value == v {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("parameter must be one of: %v", constraint.Enum)
		}
	}

	return nil
}

// addRoute adds a route to the tree
func (r *Router) addRoute(pattern string, handlers []HandlerFunc) error {
	nodes, err := parseRoute(pattern)
	if err != nil {
		return err
	}

	// Special handling for root route "/"
	if pattern == "/" {
		r.root.handlers = handlers
		r.root.pattern = "/"
		return nil
	}

	current := r.root

	for i, node := range nodes {
		node.pattern = pattern

		// Find matching child or add new one
		found := false
		for _, child := range current.children {
			if nodesMatch(child, node) {
				current = child
				found = true
				break
			}
		}

		if !found {
			current.children = append(current.children, node)
			sortChildren(current.children)
			current = node
		}

		// Set handlers on the last node
		if i == len(nodes)-1 {
			current.handlers = handlers
		}
	}

	return nil
}

// nodesMatch checks if two nodes are equivalent for merging
func nodesMatch(a, b *RouteNode) bool {
	if a.typ != b.typ {
		return false
	}

	switch a.typ {
	case RouteTypeStatic:
		return a.segment == b.segment
	case RouteTypeParam, RouteTypeTyped:
		// Parameters match if they have the same name
		if a.param != nil && b.param != nil {
			return a.param.Name == b.param.Name
		}
		return false
	case RouteTypeWildcard, RouteTypeCatchAll:
		return true
	}

	return false
}

// sortChildren sorts children by priority
func sortChildren(children []*RouteNode) {
	// Insertion sort by priority
	for i := 1; i < len(children); i++ {
		key := children[i]
		j := i - 1
		for j >= 0 && children[j].priority > key.priority {
			children[j+1] = children[j]
			j--
		}
		children[j+1] = key
	}
}

// ErrRouteNotFound indicates that no route was found for the given path
var ErrRouteNotFound = fmt.Errorf("route not found")

// findRoute finds a matching route for the given path
func (r *Router) findRoute(path string) (*MatchResult, error) {
	if path == "" || path[0] != '/' {
		return nil, fmt.Errorf("invalid path")
	}

	// Handle root route
	if path == "/" {
		if len(r.root.handlers) > 0 {
			return &MatchResult{
				Node:   r.root,
				Params: make(map[string]string),
			}, nil
		}
		return nil, ErrRouteNotFound
	}

	// Split path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	params := make(map[string]string)

	result := r.searchNode(r.root, parts, 0, params)
	if result != nil {
		return result, nil
	}

	return nil, ErrRouteNotFound
}

// searchNode recursively searches for a matching node
func (r *Router) searchNode(node *RouteNode, parts []string, index int, params map[string]string) *MatchResult {
	// Reached the end of path
	if index >= len(parts) {
		if len(node.handlers) > 0 {
			return &MatchResult{
				Node:   node,
				Params: params,
			}
		}
		return nil
	}

	// Try children in priority order
	for _, child := range node.children {
		result := r.tryMatchChild(child, parts, index, params)
		if result != nil {
			return result
		}
	}

	return nil
}

// tryMatchChild tries to match a child node
func (r *Router) tryMatchChild(node *RouteNode, parts []string, index int, params map[string]string) *MatchResult {
	currentPart := parts[index]

	switch node.typ {
	case RouteTypeStatic:
		if node.segment == currentPart {
			return r.searchNode(node, parts, index+1, params)
		}

	case RouteTypeTyped:
		if node.param != nil && node.param.Constraint != nil {
			// Validate parameter
			if err := validateParam(currentPart, node.param.Constraint); err == nil {
				// Clone params to avoid pollution
				newParams := cloneParams(params)
				newParams[node.param.Name] = currentPart
				return r.searchNode(node, parts, index+1, newParams)
			}
		}

	case RouteTypeParam:
		if node.param != nil {
			// Clone params to avoid pollution
			newParams := cloneParams(params)
			newParams[node.param.Name] = currentPart
			return r.searchNode(node, parts, index+1, newParams)
		}

	case RouteTypeWildcard:
		// * matches remaining path (like a lightweight catch-all)
		remaining := strings.Join(parts[index:], "/")
		newParams := cloneParams(params)
		newParams["*"] = remaining
		return &MatchResult{
			Node:   node,
			Params: newParams,
		}

	case RouteTypeCatchAll:
		// ** matches remaining path
		remaining := strings.Join(parts[index:], "/")
		newParams := cloneParams(params)
		newParams["*"] = remaining
		return &MatchResult{
			Node:   node,
			Params: newParams,
		}
	}

	return nil
}

// cloneParams creates a copy of params map
func cloneParams(params map[string]string) map[string]string {
	newParams := make(map[string]string, len(params))
	for k, v := range params {
		newParams[k] = v
	}
	return newParams
}
