package constants

import "fmt"

/*
* System defined types that are constants.
*
 */

type DataFieldGroup string

func (t DataFieldGroup) String() string {
	return string(t)
}

func NewDataFieldGroupFromString(s string) (DataFieldGroup, error) {
	field := DataFieldGroup(s)
	switch field {
	case DataFieldGroupCustom:
	case DataFieldGroupPrimitive:
	default:
		return "", fmt.Errorf("invalid data field group: %v", s)
	}
	return field, nil
}

type DataFieldType string

func (t DataFieldType) String() string {
	return string(t)
}

func NewDataFieldTypeFromString(s string) (DataFieldType, error) {
	dt := DataFieldType(s)
	switch dt {
	case DataFieldTypeString:
	case DataFieldTypeFloat:
	case DataFieldTypeBool:
	case DataFieldTypeInteger:
	default:
		return "", fmt.Errorf("invalid data field type: %v", s)
	}
	return dt, nil
}

func NewEntityTypeFromString(s string) (ActivityEntity, error) {
	e := ActivityEntity(s)
	switch e {
	case ActivityEntityApp:
	case ActivityEntityFlow:
	case ActivityEntityForm:
	case ActivityEntityDataType:
	case ActivityEntityDataTypeField:
	default:
		return "", fmt.Errorf("invalid entity type: %v", s)
	}
	return e, nil
}

type ActivityActionType string

func (t ActivityActionType) String() string {
	return string(t)
}

type ActivityEntity string

func (t ActivityEntity) String() string {
	return string(t)
}
