// Package client provides a REST client for the Argo Workflows API.
//
// It handles authentication, request formatting, and response parsing
// for workflow operations (create, get, list, delete, lint).
//
// The client is decoupled from the builder types in the parent [forge] package,
// following the principle that I/O concerns should be separate from pure
// data transformation.
//
// # Example
//
//	svc := client.NewWorkflowsService("https://argo.example.com", token, "default")
//	result, err := svc.CreateWorkflow(ctx, myWorkflow)
package client
