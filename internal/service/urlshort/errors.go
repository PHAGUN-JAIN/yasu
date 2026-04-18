package service

import "errors"

var ErrNotFound = errors.New("short URL not found or expired")
var ErrInvalidRequest = errors.New("invalid request")
var ErrExpired = errors.New("short URL has expired")
var ErrInternal = errors.New("internal server error")
