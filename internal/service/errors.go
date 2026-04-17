package service

import "errors"

var ErrNotFound = errors.New("short URL not found or expired")