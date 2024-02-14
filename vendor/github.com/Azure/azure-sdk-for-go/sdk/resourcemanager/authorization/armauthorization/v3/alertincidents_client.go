//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator. DO NOT EDIT.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package armauthorization

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
	"strings"
)

// AlertIncidentsClient contains the methods for the AlertIncidents group.
// Don't use this type directly, use NewAlertIncidentsClient() instead.
type AlertIncidentsClient struct {
	internal *arm.Client
}

// NewAlertIncidentsClient creates a new instance of AlertIncidentsClient with the specified values.
//   - credential - used to authorize requests. Usually a credential from azidentity.
//   - options - pass nil to accept the default values.
func NewAlertIncidentsClient(credential azcore.TokenCredential, options *arm.ClientOptions) (*AlertIncidentsClient, error) {
	cl, err := arm.NewClient(moduleName, moduleVersion, credential, options)
	if err != nil {
		return nil, err
	}
	client := &AlertIncidentsClient{
		internal: cl,
	}
	return client, nil
}

// Get - Get the specified alert incident.
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2022-08-01-preview
//   - scope - The scope of the alert incident. The scope can be any REST resource instance. For example, use '/providers/Microsoft.Subscription/subscriptions/{subscription-id}/'
//     for a subscription,
//     '/providers/Microsoft.Subscription/subscriptions/{subscription-id}/resourceGroups/{resource-group-name}' for a resource
//     group, and
//     '/providers/Microsoft.Subscription/subscriptions/{subscription-id}/resourceGroups/{resource-group-name}/providers/{resource-provider}/{resource-type}/{resource-name}'
//     for a resource.
//   - alertID - The name of the alert.
//   - alertIncidentID - The name of the alert incident to get.
//   - options - AlertIncidentsClientGetOptions contains the optional parameters for the AlertIncidentsClient.Get method.
func (client *AlertIncidentsClient) Get(ctx context.Context, scope string, alertID string, alertIncidentID string, options *AlertIncidentsClientGetOptions) (AlertIncidentsClientGetResponse, error) {
	var err error
	const operationName = "AlertIncidentsClient.Get"
	ctx = context.WithValue(ctx, runtime.CtxAPINameKey{}, operationName)
	ctx, endSpan := runtime.StartSpan(ctx, operationName, client.internal.Tracer(), nil)
	defer func() { endSpan(err) }()
	req, err := client.getCreateRequest(ctx, scope, alertID, alertIncidentID, options)
	if err != nil {
		return AlertIncidentsClientGetResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return AlertIncidentsClientGetResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK) {
		err = runtime.NewResponseError(httpResp)
		return AlertIncidentsClientGetResponse{}, err
	}
	resp, err := client.getHandleResponse(httpResp)
	return resp, err
}

// getCreateRequest creates the Get request.
func (client *AlertIncidentsClient) getCreateRequest(ctx context.Context, scope string, alertID string, alertIncidentID string, options *AlertIncidentsClientGetOptions) (*policy.Request, error) {
	urlPath := "/{scope}/providers/Microsoft.Authorization/roleManagementAlerts/{alertId}/alertIncidents/{alertIncidentId}"
	urlPath = strings.ReplaceAll(urlPath, "{scope}", scope)
	urlPath = strings.ReplaceAll(urlPath, "{alertId}", alertID)
	urlPath = strings.ReplaceAll(urlPath, "{alertIncidentId}", alertIncidentID)
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-08-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *AlertIncidentsClient) getHandleResponse(resp *http.Response) (AlertIncidentsClientGetResponse, error) {
	result := AlertIncidentsClientGetResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.AlertIncident); err != nil {
		return AlertIncidentsClientGetResponse{}, err
	}
	return result, nil
}

// NewListForScopePager - Gets alert incidents for a resource scope.
//
// Generated from API version 2022-08-01-preview
//   - scope - The scope of the alert incident.
//   - alertID - The name of the alert.
//   - options - AlertIncidentsClientListForScopeOptions contains the optional parameters for the AlertIncidentsClient.NewListForScopePager
//     method.
func (client *AlertIncidentsClient) NewListForScopePager(scope string, alertID string, options *AlertIncidentsClientListForScopeOptions) *runtime.Pager[AlertIncidentsClientListForScopeResponse] {
	return runtime.NewPager(runtime.PagingHandler[AlertIncidentsClientListForScopeResponse]{
		More: func(page AlertIncidentsClientListForScopeResponse) bool {
			return page.NextLink != nil && len(*page.NextLink) > 0
		},
		Fetcher: func(ctx context.Context, page *AlertIncidentsClientListForScopeResponse) (AlertIncidentsClientListForScopeResponse, error) {
			ctx = context.WithValue(ctx, runtime.CtxAPINameKey{}, "AlertIncidentsClient.NewListForScopePager")
			nextLink := ""
			if page != nil {
				nextLink = *page.NextLink
			}
			resp, err := runtime.FetcherForNextLink(ctx, client.internal.Pipeline(), nextLink, func(ctx context.Context) (*policy.Request, error) {
				return client.listForScopeCreateRequest(ctx, scope, alertID, options)
			}, nil)
			if err != nil {
				return AlertIncidentsClientListForScopeResponse{}, err
			}
			return client.listForScopeHandleResponse(resp)
		},
		Tracer: client.internal.Tracer(),
	})
}

// listForScopeCreateRequest creates the ListForScope request.
func (client *AlertIncidentsClient) listForScopeCreateRequest(ctx context.Context, scope string, alertID string, options *AlertIncidentsClientListForScopeOptions) (*policy.Request, error) {
	urlPath := "/{scope}/providers/Microsoft.Authorization/roleManagementAlerts/{alertId}/alertIncidents"
	urlPath = strings.ReplaceAll(urlPath, "{scope}", scope)
	urlPath = strings.ReplaceAll(urlPath, "{alertId}", alertID)
	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-08-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}

// listForScopeHandleResponse handles the ListForScope response.
func (client *AlertIncidentsClient) listForScopeHandleResponse(resp *http.Response) (AlertIncidentsClientListForScopeResponse, error) {
	result := AlertIncidentsClientListForScopeResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.AlertIncidentListResult); err != nil {
		return AlertIncidentsClientListForScopeResponse{}, err
	}
	return result, nil
}

// Remediate - Remediate an alert incident.
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2022-08-01-preview
//   - scope - The scope of the alert incident.
//   - alertID - The name of the alert.
//   - alertIncidentID - The name of the alert incident to remediate.
//   - options - AlertIncidentsClientRemediateOptions contains the optional parameters for the AlertIncidentsClient.Remediate
//     method.
func (client *AlertIncidentsClient) Remediate(ctx context.Context, scope string, alertID string, alertIncidentID string, options *AlertIncidentsClientRemediateOptions) (AlertIncidentsClientRemediateResponse, error) {
	var err error
	const operationName = "AlertIncidentsClient.Remediate"
	ctx = context.WithValue(ctx, runtime.CtxAPINameKey{}, operationName)
	ctx, endSpan := runtime.StartSpan(ctx, operationName, client.internal.Tracer(), nil)
	defer func() { endSpan(err) }()
	req, err := client.remediateCreateRequest(ctx, scope, alertID, alertIncidentID, options)
	if err != nil {
		return AlertIncidentsClientRemediateResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return AlertIncidentsClientRemediateResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusNoContent) {
		err = runtime.NewResponseError(httpResp)
		return AlertIncidentsClientRemediateResponse{}, err
	}
	return AlertIncidentsClientRemediateResponse{}, nil
}

// remediateCreateRequest creates the Remediate request.
func (client *AlertIncidentsClient) remediateCreateRequest(ctx context.Context, scope string, alertID string, alertIncidentID string, options *AlertIncidentsClientRemediateOptions) (*policy.Request, error) {
	urlPath := "/{scope}/providers/Microsoft.Authorization/roleManagementAlerts/{alertId}/alertIncidents/{alertIncidentId}/remediate"
	urlPath = strings.ReplaceAll(urlPath, "{scope}", scope)
	urlPath = strings.ReplaceAll(urlPath, "{alertId}", alertID)
	urlPath = strings.ReplaceAll(urlPath, "{alertIncidentId}", alertIncidentID)
	req, err := runtime.NewRequest(ctx, http.MethodPost, runtime.JoinPaths(client.internal.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2022-08-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}