// Package openapi contains the documented Pixels HTTP route contract.
package openapi

// Spec is the OpenAPI document for the HTTP surface.
const Spec = `{
  "openapi": "3.1.0",
  "info": {
    "title": "Pixels API",
    "version": "0.1.0",
    "description": "Public status, websocket entrypoint, development documentation, and protected emulator endpoints."
  },
  "servers": [
    {
      "url": "http://{host}:{port}",
      "variables": {
        "host": { "default": "127.0.0.1" },
        "port": { "default": "3000" }
      }
    }
  ],
  "components": {
    "parameters": {
      "ApiKeyHeader": {
        "name": "X-API-Key",
        "in": "header",
        "required": true,
        "schema": { "type": "string" },
        "description": "Access key configured by PIXELS_ACCESS_KEY."
      },
      "ConnectionHeader": {
        "name": "Connection",
        "in": "header",
        "required": true,
        "schema": { "type": "string", "example": "Upgrade" },
        "description": "Required websocket connection upgrade header."
      },
      "UpgradeHeader": {
        "name": "Upgrade",
        "in": "header",
        "required": true,
        "schema": { "type": "string", "example": "websocket" },
        "description": "Required websocket upgrade header."
      }
    },
    "securitySchemes": {
      "ApiKeyAuth": {
        "type": "apiKey",
        "in": "header",
        "name": "X-API-Key"
      }
    },
    "schemas": {
      "ErrorResponse": {
        "type": "object",
        "required": ["error"],
        "properties": {
          "error": { "type": "string" }
        }
      },
      "StatusResponse": {
        "type": "object",
        "required": ["status", "environment", "version"],
        "properties": {
          "status": { "type": "string", "example": "ok" },
          "environment": { "type": "string", "example": "development" },
          "version": { "type": "string", "example": "0.1.0-abcdef12" }
        }
      },
      "CreateSSOTicketRequest": {
        "type": "object",
        "required": ["userId"],
        "properties": {
          "userId": { "type": "string", "description": "Temporary TODO user id bound to the SSO ticket." },
          "ip": { "type": "string", "description": "Optional IP address allowed to consume the ticket." },
          "ttlSeconds": { "type": "integer", "minimum": 1, "description": "Optional TTL override in seconds." }
        }
      },
      "CreateSSOTicketResponse": {
        "type": "object",
        "required": ["ticket", "expiresAt"],
        "properties": {
          "ticket": { "type": "string", "description": "Opaque one-time SSO ticket." },
          "expiresAt": { "type": "string", "format": "date-time" }
        }
      }
    }
  },
  "paths": {
    "/status": {
      "get": {
        "summary": "Read server status",
        "description": "Returns public runtime status without requiring an API key.",
        "responses": {
          "200": {
            "description": "Server status.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/StatusResponse" }
              }
            }
          }
        }
      }
    },
    "/ws": {
      "get": {
        "summary": "Open websocket session",
        "description": "Upgrades an HTTP request to the pixel-protocol websocket entrypoint.",
        "parameters": [
          { "$ref": "#/components/parameters/ConnectionHeader" },
          { "$ref": "#/components/parameters/UpgradeHeader" }
        ],
        "responses": {
          "101": { "description": "Websocket upgrade accepted." },
          "426": {
            "description": "Upgrade header is missing or invalid.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
                "example": { "error": "websocket upgrade required" }
              }
            }
          }
        }
      }
    },
    "/docs": {
      "get": {
        "summary": "Read Scalar API documentation",
        "description": "Serves public Scalar documentation in development only.",
        "responses": {
          "200": {
            "description": "Scalar documentation HTML.",
            "content": {
              "text/html": {
                "schema": { "type": "string" }
              }
            }
          },
          "404": {
            "description": "Documentation is disabled outside development.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
                "example": { "error": "not found" }
              }
            }
          }
        }
      }
    },
    "/api/sso/tickets": {
      "post": {
        "summary": "Create SSO ticket",
        "description": "Creates a Redis-backed one-time SSO ticket for the configured TTL.",
        "security": [{ "ApiKeyAuth": [] }],
        "parameters": [
          { "$ref": "#/components/parameters/ApiKeyHeader" }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CreateSSOTicketRequest" }
            }
          }
        },
        "responses": {
          "201": {
            "description": "SSO ticket created.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/CreateSSOTicketResponse" }
              }
            }
          },
          "400": {
            "description": "Request body is invalid or missing userId.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
                "example": { "error": "Bad Request" }
              }
            }
          },
          "401": {
            "description": "API key is missing or invalid.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
                "example": { "error": "unauthorized" }
              }
            }
          },
          "500": {
            "description": "Redis or ticket storage failed.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorResponse" }
              }
            }
          }
        }
      }
    },
    "/*": {
      "get": {
        "summary": "Private route fallback",
        "description": "Represents protected endpoints added after public route registration.",
        "security": [{ "ApiKeyAuth": [] }],
        "parameters": [
          { "$ref": "#/components/parameters/ApiKeyHeader" }
        ],
        "responses": {
          "401": {
            "description": "API key is missing or invalid.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
                "example": { "error": "unauthorized" }
              }
            }
          },
          "404": {
            "description": "Authenticated route was not found.",
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/ErrorResponse" },
                "example": { "error": "not found" }
              }
            }
          }
        }
      }
    }
  }
}`

// Bytes returns the OpenAPI document bytes.
func Bytes() []byte {
	return []byte(Spec)
}
