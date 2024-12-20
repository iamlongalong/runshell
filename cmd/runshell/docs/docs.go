// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/commands": {
            "get": {
                "description": "List all available commands",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "commands"
                ],
                "summary": "List Commands",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.CommandInfo"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/exec": {
            "post": {
                "description": "Execute a shell command",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "commands"
                ],
                "summary": "Execute Command",
                "parameters": [
                    {
                        "description": "Command execution request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ExecRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ExecResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/health": {
            "get": {
                "description": "Check if the server is running",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Health Check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/help": {
            "get": {
                "description": "Get help information for a specific command",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "commands"
                ],
                "summary": "Get Command Help",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Command name",
                        "name": "command",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/sessions": {
            "get": {
                "description": "List all active sessions",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sessions"
                ],
                "summary": "List Sessions",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.Session"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new session",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sessions"
                ],
                "summary": "Create Session",
                "parameters": [
                    {
                        "description": "Session creation request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.SessionRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.SessionResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/sessions/{id}": {
            "delete": {
                "description": "Delete a session by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sessions"
                ],
                "summary": "Delete Session",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/sessions/{id}/exec": {
            "post": {
                "description": "Execute a command in a specific session",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sessions"
                ],
                "summary": "Execute Command in Session",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Session ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Command execution request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ExecRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ExecResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/pkg_server.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_iamlongalong_runshell_pkg_types.CommandInfo": {
            "type": "object",
            "properties": {
                "category": {
                    "description": "命令类",
                    "type": "string"
                },
                "description": {
                    "description": "命令描述",
                    "type": "string",
                    "example": "List directory"
                },
                "metadata": {
                    "description": "命令元数据",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "name": {
                    "description": "命令名称",
                    "type": "string",
                    "example": "ls"
                },
                "usage": {
                    "description": "命令用法",
                    "type": "string",
                    "example": "ls [options] [path]"
                }
            }
        },
        "github_com_iamlongalong_runshell_pkg_types.DockerConfig": {
            "type": "object",
            "properties": {
                "allowUnregisteredCommands": {
                    "description": "是否允许执行未注册的命令",
                    "type": "boolean"
                },
                "bindMount": {
                    "description": "目录绑定",
                    "type": "string"
                },
                "image": {
                    "description": "Docker 镜像",
                    "type": "string"
                },
                "useBuiltinCommands": {
                    "description": "是否使用内置命令",
                    "type": "boolean"
                },
                "user": {
                    "description": "用户",
                    "type": "string"
                },
                "workDir": {
                    "description": "工作目录",
                    "type": "string"
                }
            }
        },
        "github_com_iamlongalong_runshell_pkg_types.ExecuteOptions": {
            "type": "object",
            "properties": {
                "env": {
                    "description": "Env 指定命令执行时的环境变量",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "metadata": {
                    "description": "Metadata 存储额外的元数据信息",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "shell": {
                    "description": "Shell 指定执行命令的 shell, 默认使用 /bin/bash",
                    "type": "string"
                },
                "timeout": {
                    "description": "Timeout 指定命令执行的超时时间（纳秒）\nswagger:strfmt int64",
                    "type": "integer",
                    "example": 30000000000
                },
                "tty": {
                    "description": "TTY 是否分配伪终端",
                    "type": "boolean"
                },
                "user": {
                    "description": "User 指定执行命令的用户信息",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.User"
                        }
                    ]
                },
                "workdir": {
                    "description": "WorkDir 指定命令执行的工作目录，所有相对路径都相对于此目录",
                    "type": "string"
                }
            }
        },
        "github_com_iamlongalong_runshell_pkg_types.LocalConfig": {
            "type": "object",
            "properties": {
                "allowUnregisteredCommands": {
                    "description": "是否允许执行未注册的命令",
                    "type": "boolean"
                },
                "useBuiltinCommands": {
                    "description": "是否使用内置命令",
                    "type": "boolean"
                },
                "workDir": {
                    "description": "工作目录",
                    "type": "string"
                }
            }
        },
        "github_com_iamlongalong_runshell_pkg_types.Session": {
            "type": "object",
            "properties": {
                "created_at": {
                    "description": "会话创建时间",
                    "type": "string"
                },
                "id": {
                    "description": "会话的唯一标识符",
                    "type": "string",
                    "example": "sess_123"
                },
                "last_accessed_at": {
                    "description": "最后访问时间",
                    "type": "string"
                },
                "metadata": {
                    "description": "会话相关的元数据",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "options": {
                    "description": "会话的执行选项",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.ExecuteOptions"
                        }
                    ]
                },
                "status": {
                    "description": "会话状态",
                    "type": "string"
                }
            }
        },
        "github_com_iamlongalong_runshell_pkg_types.SessionRequest": {
            "type": "object",
            "properties": {
                "docker_config": {
                    "description": "Docker 执行器配置",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.DockerConfig"
                        }
                    ]
                },
                "executor_type": {
                    "description": "执行器类型（local/docker）",
                    "type": "string"
                },
                "local_config": {
                    "description": "本地执行器配置",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.LocalConfig"
                        }
                    ]
                },
                "metadata": {
                    "description": "会话元数据",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "options": {
                    "description": "执行选项",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.ExecuteOptions"
                        }
                    ]
                }
            }
        },
        "github_com_iamlongalong_runshell_pkg_types.SessionResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "错误信息",
                    "type": "string"
                },
                "session": {
                    "description": "会话信息",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_iamlongalong_runshell_pkg_types.Session"
                        }
                    ]
                }
            }
        },
        "github_com_iamlongalong_runshell_pkg_types.User": {
            "type": "object",
            "properties": {
                "gid": {
                    "description": "GID 是用户组 ID",
                    "type": "integer"
                },
                "groups": {
                    "description": "Groups 是用户所属的附加组 ID 列表",
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "uid": {
                    "description": "UID 是用户 ID",
                    "type": "integer"
                },
                "username": {
                    "description": "Username 是用户名",
                    "type": "string"
                }
            }
        },
        "pkg_server.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "错误信息",
                    "type": "string",
                    "example": "Invalid request parameter"
                }
            }
        },
        "pkg_server.ExecRequest": {
            "type": "object",
            "required": [
                "command"
            ],
            "properties": {
                "args": {
                    "description": "命令参数",
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "[\"-l\"",
                        "\"-a\"]"
                    ]
                },
                "command": {
                    "description": "要执行的命令",
                    "type": "string",
                    "example": "ls"
                },
                "env": {
                    "description": "环境变量",
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "workdir": {
                    "description": "工作目录",
                    "type": "string"
                }
            }
        },
        "pkg_server.ExecResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "description": "错误信息，如果有的话",
                    "type": "string"
                },
                "exit_code": {
                    "description": "命令退出码",
                    "type": "integer",
                    "example": 0
                },
                "output": {
                    "description": "命令输出",
                    "type": "string",
                    "example": "file1.txt"
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http"},
	Title:            "RunShell API",
	Description:      "API for executing and managing shell commands",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
