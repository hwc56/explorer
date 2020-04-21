// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2020-04-21 10:22:03.129157 +0800 CST m=+0.136114813

package docs

import (
	"bytes"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "info": {
        "contact": {},
        "license": {}
    },
    "paths": {
        "/api/config": {
            "get": {
                "description": "get config",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "common"
                ],
                "summary": "config",
                "responses": {
                    "200": {
                        "description": "success",
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/vo.EnvConfig"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "vo.ConfigVo": {
            "type": "object",
            "properties": {
                "chain_id": {
                    "type": "string"
                },
                "env": {
                    "type": "string"
                },
                "host": {
                    "type": "string"
                },
                "network_name": {
                    "type": "string"
                },
                "node_version": {
                    "type": "string"
                },
                "show_faucet": {
                    "type": "integer"
                },
                "tendermint_version": {
                    "type": "string"
                },
                "umeng_id": {
                    "type": "integer"
                }
            }
        },
        "vo.EnvConfig": {
            "type": "object",
            "properties": {
                "chain_id": {
                    "type": "string"
                },
                "configs": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/vo.ConfigVo"
                    }
                },
                "cur_env": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo swaggerInfo

type s struct{}

func (s *s) ReadDoc() string {
	t, err := template.New("swagger_info").Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, SwaggerInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
