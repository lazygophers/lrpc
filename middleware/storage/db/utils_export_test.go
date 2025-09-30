package db

import "reflect"

// Export private functions for testing

func ExportDecode(field reflect.Value, col []byte) error {
	return decode(field, col)
}

func ExportScanComplexType(field reflect.Value, col []byte, isPtr bool) error {
	return scanComplexType(field, col, isPtr)
}

func ExportGetTableName(elem reflect.Type) string {
	return getTableName(elem)
}

func ExportHasField(elem reflect.Type, fieldName string) bool {
	return hasField(elem, fieldName)
}