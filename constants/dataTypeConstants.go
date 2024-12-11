package constants

const (
	DataFieldGroupCustom    = DataFieldGroup("CUSTOM")
	DataFieldGroupPrimitive = DataFieldGroup("PRIMITIVE")
)

const (
	DataFieldTypeString  = DataFieldType("STRING")
	DataFieldTypeInteger = DataFieldType("INTEGER")
	DataFieldTypeFloat   = DataFieldType("FLOAT")
	DataFieldTypeBool    = DataFieldType("BOOLEAN")
)

const (
	ActivityActionTypeAdd    = ActivityActionType("ADD")
	ActivityActionTypeRemove = ActivityActionType("REMOVE")
	ActivityActionTypeEdit   = ActivityActionType("EDIT")
)
const (
	ActivityEntityApp           = ActivityEntity("APP")
	ActivityEntityFlow          = ActivityEntity("FLOW")
	ActivityEntityForm          = ActivityEntity("FORM")
	ActivityEntityDataType      = ActivityEntity("DATATYPE")
	ActivityEntityDataTypeField = ActivityEntity("DATATYPEFIELD")
)

const (
	TokenTypePermanent = "PERMANENT"
	TokenTypeTemporary = "TEMPORARY"
)
