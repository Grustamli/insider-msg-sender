// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "consumes": [
        "application/json"
    ],
    "produces": [
        "application/json"
    ],
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Gadir Rustamli",
            "email": "gadir.rustamli@outlook.com"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/messages": {
            "get": {
                "description": "Retrieve all messages that have been sent, including their IDs and timestamps.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Scheduler"
                ],
                "summary": "List sent messages",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.ListSentMessagesResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/start": {
            "post": {
                "description": "Initiates the scheduler to begin sending messages at configured intervals.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Scheduler"
                ],
                "summary": "Start message sender",
                "operationId": "startSender",
                "responses": {
                    "202": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/stop": {
            "post": {
                "description": "Halts the scheduler, stopping any further message dispatch until restarted.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Scheduler"
                ],
                "summary": "Stop the message sender",
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.ListSentMessagesResponse": {
            "type": "object",
            "properties": {
                "items": {
                    "description": "items is the array of messages that have been sent.",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/api.MessageOut"
                    }
                }
            }
        },
        "api.MessageOut": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "sent_at": {
                    "type": "string"
                }
            }
        }
    },
    "tags": [
        {
            "name": "Scheduler"
        }
    ]
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8000",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Insider Message Sender API",
	Description:      "API endpoints for the Insider Message Sender that periodically sends messages from DB",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
