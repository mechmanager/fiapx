// Package domain contém as entidades centrais e as interfaces (ports) do api-gateway.
package domain

import "errors"

// Erros de domínio reutilizados pelas camadas de serviço e handler.
var (
	// ErrEmailAlreadyExists indica tentativa de cadastro com email já usado.
	ErrEmailAlreadyExists = errors.New("email já cadastrado")
	// ErrInvalidCredentials indica email ou senha inválidos no login.
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	// ErrUserNotFound indica que o usuário não foi encontrado.
	ErrUserNotFound = errors.New("usuário não encontrado")
	// ErrVideoNotFound indica que o vídeo não foi encontrado.
	ErrVideoNotFound = errors.New("vídeo não encontrado")
	// ErrVideoNotReady indica que o zip do vídeo ainda não está disponível.
	ErrVideoNotReady = errors.New("vídeo ainda não foi processado")
	// ErrInvalidVideoFormat indica extensão de arquivo não suportada.
	ErrInvalidVideoFormat = errors.New("formato de vídeo não suportado")
)
