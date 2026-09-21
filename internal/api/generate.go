// Package api chứa type sinh ra từ openapi.yaml — spec là nguồn sự thật duy
// nhất của contract. Không sửa tay: chạy `make openapi` sau khi đổi spec.
package api

//go:generate oapi-codegen -config ../../oapi-codegen.yaml ../../openapi.yaml
