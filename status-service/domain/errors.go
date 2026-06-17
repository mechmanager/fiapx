package domain

import "errors"

// ErrVideoNotFound é retornado quando o vídeo não existe no banco.
var ErrVideoNotFound = errors.New("vídeo não encontrado")

// ErrVideoNotReady é retornado quando o vídeo ainda não foi processado.
var ErrVideoNotReady = errors.New("vídeo ainda não foi processado")
