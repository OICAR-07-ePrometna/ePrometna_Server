// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Authenticates a user and returns access and refresh tokens",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "User login",
                "parameters": [
                    {
                        "description": "Login credentials",
                        "name": "loginDto",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.LoginDto"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.TokenDto"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Generates a new access token using a valid refresh token",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Refresh Access Token",
                "parameters": [
                    {
                        "description": "Refresh Token",
                        "name": "refreshToken",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.TokenDto"
                        }
                    }
                }
            }
        },
        "/user": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Create new user",
                "parameters": [
                    {
                        "description": "Data for new user",
                        "name": "model",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.NewUserDto"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/dto.UserDto"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user/all-users": {
            "get": {
                "description": "Fetches all users for superadmin",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get all users for superadmin",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.UserDto"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user/my-data": {
            "get": {
                "description": "Fetches the currently logged-in user's data based on the JWT token",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get logged-in user data",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.UserDto"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user/police-officers": {
            "get": {
                "description": "Fetches all police officers for MUP Admin",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get all police officers for MUP Admin",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.UserDto"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user/search": {
            "get": {
                "description": "Performs a fuzzy search for users by first name, last name, or full name with similarity matching",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Search users by name",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Search query",
                        "name": "query",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.UserDto"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/user/{uuid}": {
            "get": {
                "description": "get a user with uuid",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "get user with uuid",
                "parameters": [
                    {
                        "type": "string",
                        "description": "user uuid",
                        "name": "uuid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.UserDto"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "put": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Update user with new dat",
                "parameters": [
                    {
                        "type": "string",
                        "description": "uuid of user to be updated",
                        "name": "uuid",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Data for updating user",
                        "name": "model",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.UserDto"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.UserDto"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "delete": {
                "description": "delete a user with uuid",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "delete user with uuid",
                "parameters": [
                    {
                        "type": "string",
                        "description": "user uuid",
                        "name": "uuid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/vehicle": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vehicle"
                ],
                "summary": "Gets your vehicles",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.VehicleDto"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "post": {
                "description": "Create new vehicle with an owner",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vehicle"
                ],
                "summary": "Creates new vehicle",
                "parameters": [
                    {
                        "description": "Vehicle model",
                        "name": "model",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.NewVehicleDto"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/dto.VehicleDto"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/vehicle/{uuid}": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vehicle"
                ],
                "summary": "Gets a vehicle with uuid",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Vehicle UUID",
                        "name": "uuid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/dto.VehicleDetailsDto"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            },
            "delete": {
                "description": "Preforms a soft delete",
                "tags": [
                    "vehicle"
                ],
                "summary": "Soft delete on vehicle",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Vehicle UUID",
                        "name": "uuid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        }
    },
    "definitions": {
        "dto.LoginDto": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "minLength": 6
                }
            }
        },
        "dto.NewUserDto": {
            "type": "object",
            "required": [
                "birthDate",
                "email",
                "firstName",
                "lastName",
                "oib",
                "password",
                "residence",
                "role"
            ],
            "properties": {
                "birthDate": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "firstName": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 2
                },
                "lastName": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 2
                },
                "oib": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "minLength": 6
                },
                "residence": {
                    "type": "string",
                    "maxLength": 255
                },
                "role": {
                    "type": "string",
                    "enum": [
                        "hak",
                        "mupadmin",
                        "osoba",
                        "firma",
                        "policija",
                        "superadmin"
                    ]
                },
                "uuid": {
                    "type": "string"
                }
            }
        },
        "dto.NewVehicleDto": {
            "type": "object",
            "properties": {
                "chassisNumber": {
                    "type": "string"
                },
                "ownerUuid": {
                    "type": "string"
                },
                "productionYear": {
                    "type": "integer"
                },
                "registration": {
                    "type": "string"
                },
                "traveledDistance": {
                    "type": "integer"
                },
                "vehicleModel": {
                    "type": "string"
                },
                "vehicleType": {
                    "type": "string"
                }
            }
        },
        "dto.TokenDto": {
            "type": "object",
            "properties": {
                "accessToken": {
                    "type": "string"
                },
                "refreshToken": {
                    "type": "string"
                }
            }
        },
        "dto.UserDto": {
            "type": "object",
            "properties": {
                "birthDate": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "firstName": {
                    "type": "string"
                },
                "lastName": {
                    "type": "string"
                },
                "oib": {
                    "type": "string"
                },
                "residence": {
                    "type": "string"
                },
                "role": {
                    "type": "string"
                },
                "uuid": {
                    "type": "string"
                }
            }
        },
        "dto.VehicleDetailsDto": {
            "type": "object",
            "properties": {
                "drivers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.UserDto"
                    }
                },
                "owner": {
                    "$ref": "#/definitions/dto.UserDto"
                },
                "pastOwners": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/dto.UserDto"
                    }
                },
                "productionYear": {
                    "type": "integer"
                },
                "registration": {
                    "type": "string"
                },
                "uuid": {
                    "type": "string"
                },
                "vehicleModel": {
                    "type": "string"
                },
                "vehicleType": {
                    "type": "string"
                }
            }
        },
        "dto.VehicleDto": {
            "type": "object",
            "properties": {
                "productionYear": {
                    "type": "integer"
                },
                "registration": {
                    "type": "string"
                },
                "uuid": {
                    "type": "string"
                },
                "vehicleModel": {
                    "type": "string"
                },
                "vehicleType": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
