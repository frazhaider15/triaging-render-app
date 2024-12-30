package models

import (
	"time"

	"gorm.io/gorm"
)

type Flow struct {
	ID          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	BuilderJson Workflow       `json:"builder_json"`
	AppId       int            `json:"app_id"`
}

type Workflow struct {
	FlowWorkspace struct {
		Edges   []Edge `json:"edges"`
		Nodes   []Node `json:"nodes"`
		Details struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"details"`
		EdgesData map[string]map[string]string `json:"edgesData"`
		NodesData map[string]interface{}       `json:"nodesData"`
	} `json:"flowWorkspace"`
}

type Edge struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Source       string      `json:"source"`
	Target       string      `json:"target"`
	Animated     string      `json:"animated"`
	SourceHandle interface{} `json:"sourceHandle"`
	TargetHandle interface{} `json:"targetHandle"`
	Selected     bool        `json:"selected"`
}

type Node struct {
	ID   string `json:"id"`
	Data struct {
		Label string `json:"label"`
	} `json:"data"`
	Type     string `json:"type"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Dragging bool   `json:"dragging"`
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position"`
	Selected         bool `json:"selected"`
	PositionAbsolute struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
}

type ApiNodeData struct {
	API struct {
		Body        string      `json:"body"`
		Description string      `json:"description"`
		MapType     string      `json:"mapType"`
		Headers     []ApiHeader `json:"headers"`
		Name        string      `json:"name"`
		QueryParams []ApiHeader `json:"queryParams"`
		ReqMethod   string      `json:"reqMethod"`
		URL         string      `json:"url"`
	} `json:"api"`
}

type FormNodeData struct {
	FormId int `json:"formId"`
}

type ApiHeader struct {
	ID        string `json:"id"`
	KeyItem   string `json:"keyItem"`
	ValueItem string `json:"valueItem"`
}

type ConditionNodeData struct {
	Logic struct {
		Conditions []struct {
			Name  string `json:"name"`
			Rules []struct {
				ID              string      `json:"id"`
				Field           string      `json:"field"`
				Value           string      `json:"value"`
				OperandA        interface{} `json:"operandA"`
				OperandB        interface{} `json:"operandB"`
				Operator        string      `json:"operator"`
				LogicalOperator interface{} `json:"logicalOperator"`
			} `json:"rules"`
		} `json:"conditions"`
	} `json:"logic"`
}

type DataDictionaryError struct {
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    string `json:"errorCode"`
}

type Tag struct {
	gorm.Model

	Name     string
	Colour   string
	TenantId string
}

type EntityTags struct {
	gorm.Model

	TagId      uint
	EntityId   uint
	EntityType string
}

type Form struct {
	ID          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	BuilderJson interface{}    `json:"builder_json"`
	AppId       int            `json:"app_id"`
	Tags        []byte         `json:"tags,omitempty"`
	Type        string         `json:"type"`
}

type Theme struct {
	ID          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Theme       interface{}    `json:"theme"`
	AppId       int            `json:"app_id"`
	IsDefault   bool           `json:"is_default"`
}

type FlowTheme struct {
	ID        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
	FlowId    string         `json:"flow_id"`
	ThemeId   string         `json:"theme_id"`
}

type DataType struct {
	ID        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
	Name      string         `json:"name"`
	AppId     uint           `json:"app_id"`
}

type DataTypeWithFields struct {
	ID        uint            `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `json:"deleted_at"`
	Name      string          `json:"name"`
	AppId     uint            `json:"app_id"`
	Fields    []DataTypeField `json:"data_type_fields"`
}

type DataTypeField struct {
	ID         uint           `json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	IsArray    bool           `json:"is_array"`
	TypeGroup  string         `json:"type_group"`
	DataTypeId uint           `json:"data_type_id"`
}

type App struct {
	ID          uint           `json:"app_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	UserId      string         `json:"user_id"`
	TenantId    string         `json:"tenant_id"`
	Details     AppDetails     `json:"details"`
}
type AppDetails struct {
	Flows      []Flow               `json:"flows"`
	Forms      []Form               `json:"forms"`
	Themes     []Theme              `json:"themes"`
	DataTypes  []DataTypeWithFields `json:"dataTypes"`
	AppTokens  []AppToken           `json:"app_tokens"`
	FlowThemes []FlowTheme          `json:"flow_themes"`
	PageRoutes []PageRoute          `json:"page_routes"`
}

type AppToken struct {
	ID          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
	UserId      string         `json:"user_id"`
	FlowId      uint           `json:"flow_id"`
	AppId       uint           `json:"app_id"`
	Token       string         `json:"token"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
}

type PageRoute struct {
	ID        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
	Path      string         `json:"path"`
	PageId    int            `json:"page_id"`
	AppId     int            `json:"app_id"`
}

type AppTokenById struct {
	TokenId     uint      `gorm:"column:id" json:"id"`
	UserId      string    `gorm:"column:user_id" json:"userId"`
	FlowId      uint      `gorm:"column:flow_id" json:"flowId"`
	AppId       uint      `gorm:"column:app_id" json:"appId"`
	FlowName    string    `gorm:"column:title" json:"flowName"`
	Token       string    `gorm:"column:token" json:"token"`
	Description string    `gorm:"column:description" json:"tokenDescription"`
	Creator     string    `gorm:"column:name" json:"creator"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}
