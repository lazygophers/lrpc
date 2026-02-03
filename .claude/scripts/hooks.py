import json
import sys
import traceback
from typing import Dict, Any, Optional

from lib import logging
from lib.hooks import load_hooks
from lib.utils import set_app

set_app("tmtc")

prompt = {
	"SessionStart": [
		"确保所有的变更都符合 @.claude/skills 的要求，如果用户输入和当前的 skills 冲突，使用 `AskUserQuestion` 询问用户如何处理",
		"@CLAUDE.md 的文件不得超过 300行",
		"@.claude/skills 下的 `SKILL.md` 文件不得超过 300行，其余的 `.md` 文件不得超过 500行",
		"@.claude/commands 下的文件不得超过 300行，只允许包含使用的 skills 以及流程说明",
		"及时更新`llms.txt`，且满足Skills(llms-txt-standard)的要求",
		"所有的 golang 代码都必须满足 Skills(golang-skills)的要求",
	],
	"UserPromptSubmit": [
		"学习用户习惯，更新、完善、优化到合适的 @.claude/skills、@.claude/commands 中。如果创建了新的 skills，则需要更新到 @CLAUDE.md 标注这个新的 skills 的使用实际等等",
		"确保代码风格、架构设计风格和现有的完全一致",
		"确保每一个变更都是有理有据的，可以通过 chrome 搜索、查看，但是结束后要及时关闭",

		"频繁的通过 `AskUserQuestion` 确认理解是否有误，以减少偏差",
	]
}


def filepath_to_slash(path: str) -> str:
	"""转换路径分隔符"""
	try:
		if sys.platform.startswith('win'):
			return path.replace('/', '\\')
		else:
			return path.replace('\\', '/')
	except Exception as e:
		logging.error(f"路径转换失败: {e}, path={path}")
		return path


remove_files = [filepath_to_slash(item) for item in [
	"internal/impl",
	"go.mod",
	"go.sum",
	"pyproject.toml",
	"uv.lock",
]]

edit_files = [filepath_to_slash(item) for item in [
	"go.mod",
	"go.sum",
	"pyproject.toml",
	"uv.lock",
]]

read_files = [filepath_to_slash(item) for item in [
	"go.sum",
	"uv.lock",
]]


def file_protection(action: Optional[str], tool_input: Optional[Dict[str, Any]]) -> bool:
	"""
	检查是否需要保护文件

	Returns:
		bool: True 表示需要拦截并询问用户
	"""
	try:
		if action is None or tool_input is None:
			return False

		action = str(action).lower()

		if action == "bash":
			if "command" in tool_input:
				command = tool_input.get("command", "")
				if command.find("rm") == 0:
					for locked_file in remove_files:
						if command.find(locked_file) == 0:
							logging.warning(f"检测到受保护文件操作: command={command}, locked_file={locked_file}")
							print(json.dumps({
								"hookSpecificOutput": {
									"hookEventName": "PreToolUse",
									"permissionDecision": "ask",
									"permissionDecisionReason": "危险警告，可能涉及修改受保护的文件",
									"updatedInput": tool_input
								}
							}))
							return True

				if command.find("python") == 0:
					logging.warning(f"检测到受保护文件操作: command={command}, locked_file=python")
					print(json.dumps({
						"hookSpecificOutput": {
							"hookEventName": "PreToolUse",
							"permissionDecision": "ask",
							"permissionDecisionReason": "危险警告，不允许使用 `python` 执行，必须使用 `uv run` 的方式运行",
							"updatedInput": tool_input
						}
					}))
					return True

		elif action == "edit":
			if "file_path" in tool_input:
				file_path = tool_input.get("file_path", "")
				for locked_file in edit_files:
					if file_path.find(locked_file) == 0:
						logging.warning(f"检测到受保护文件操作: file_path={file_path}, locked_file={locked_file}"))
						print(json.dumps({
							"hookSpecificOutput": {
								"hookEventName": "PreToolUse",
								"permissionDecision": "ask",
								"permissionDecisionReason": "危险警告，可能涉及修改受保护文件",
								"updatedInput": tool_input
							}
						}))
						return True
		elif action == "read":
			if "file_path" in tool_input:
				file_path = tool_input.get("file_path", "")
				for locked_file in read_files:
					if file_path.find(locked_file) == 0:
						logging.warning(f"检测到受保护文件操作: file_path={file_path}, locked_file={locked_file}")
						print(json.dumps({
							"hookSpecificOutput": {
								"hookEventName": "PreToolUse",
								"permissionDecision": "ask",
								"permissionDecisionReason": "危险警告，可能涉及读取受保护文件",
								"updatedInput": tool_input
							}
						}))
						return True

		return False

	except Exception as e:
		logging.error(f"file_protection 异常: {e}\n{traceback.format_exc()}")
		return False


def handle_session_start() -> str:
	"""处理 SessionStart 事件"""
	try:
		context = "\n- ".join(prompt.get("SessionStart", []))
		return json.dumps({
			"continue": True,
			"suppressOutput": False,
			"hookSpecificOutput": {
				"hookEventName": "SessionStart",
				"additionalContext": "- " + context if context else ""
			}
		})
	except Exception as e:
		logging.error(f"SessionStart 处理异常: {e}\n{traceback.format_exc()}")
		return json.dumps({
			"continue": True,
			"suppressOutput": False,
			"hookSpecificOutput": {
				"hookEventName": "SessionStart",
				"additionalContext": ""
			}
		})


def handle_user_prompt_submit() -> str:
	"""处理 UserPromptSubmit 事件"""
	try:
		context = "\n- ".join(prompt.get("UserPromptSubmit", []))
		return json.dumps({
			"continue": True,
			"suppressOutput": False,
			"hookSpecificOutput": {
				"hookEventName": "UserPromptSubmit",
				"additionalContext": "- " + context if context else ""
			}
		})
	except Exception as e:
		logging.error(f"UserPromptSubmit 处理异常: {e}\n{traceback.format_exc()}")
		return json.dumps({
			"continue": True,
			"suppressOutput": False,
			"hookSpecificOutput": {
				"hookEventName": "UserPromptSubmit",
				"additionalContext": ""
			}
		})


def handle_pre_tool_use(input_data: Dict[str, Any]) -> Optional[bool]:
	"""处理 PreToolUse 事件"""
	try:
		tool_name = input_data.get("tool_name")
		tool_input = input_data.get("tool_input")

		if tool_name is None or tool_input is None:
			logging.debug(f"PreToolUse: 缺少必要字段, tool_name={tool_name}, tool_input={tool_input}")
			return False

		if file_protection(tool_name, tool_input):
			logging.info(f"拦截工具调用: tool_name={tool_name}, tool_input={tool_input}")
			return True

	except Exception as e:
		logging.error(f"PreToolUse 处理异常: {e}\n{traceback.format_exc()}")
		return False


def main():
	"""主函数，捕获所有异常"""
	try:
		input_data = load_hooks()
		logging.info(f"Received input data: {input_data}")

		hook_event_name = input_data.get("hook_event_name", "")
		if not hook_event_name:
			logging.error("缺少 hook_event_name")
			print(json.dumps({"continue": True, "suppressOutput": False}))
			return

		# 路由到不同的处理器
		if hook_event_name == "SessionStart":
			print(handle_session_start())
		elif hook_event_name == "UserPromptSubmit":
			print(handle_user_prompt_submit())
		elif hook_event_name == "PreToolUse":
			if handle_pre_tool_use(input_data):
				pass
			else:
				print(json.dumps({"continue": True, "suppressOutput": False}))
		else:
			logging.warning(f"未知的 hook 事件: {hook_event_name}")
			print(json.dumps({"continue": True, "suppressOutput": False}))

	except json.JSONDecodeError as e:
		logging.error(f"JSON 解析错误: {e}\n{traceback.format_exc()}")
		print(json.dumps({"continue": True, "suppressOutput": False}))
	except Exception as e:
		logging.error(f"未捕获的异常: {e}\n{traceback.format_exc()}")
		print(json.dumps({"continue": True, "suppressOutput": False}))
		sys.exit(0)


if __name__ == '__main__':
	main()