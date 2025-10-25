package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"herois-da-pilha/data"
	"herois-da-pilha/util"

	"github.com/sashabaranov/go-openai"
)

// FinderService é o struct que gerencia a lógica de IA e o cache.
type FinderService struct {
	openAIClient *openai.Client
	modelName    string
}

// NewFinderService inicializa o cliente OpenAI (OpenRouter) e o cache.
func NewFinderService() *FinderService {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("AVISO: Variável OPENROUTER_API_KEY não está definida!")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"

	// Modelo recomendado para performance e custo
	model := "openai/gpt-4o-mini"

	s := &FinderService{
		openAIClient: openai.NewClientWithConfig(config),
		modelName:    model,
	}

	return s
}

func getPrompt() string {
	return data.IntentClassificationPrompt
}

// FindService usa o cache ou o modelo de IA para classificar a intenção.
func (s *FinderService) FindService(intent string) util.FindServiceResponse {
	// 1. TENTAR LER DO CACHE (Leitura Rápida)
	// s.mu.RLock() // Removed
	// if data, ok := s.cache[intent]; ok { // Removed
	// 	s.mu.RUnlock() // Removed
	// 	return &data, nil // Cache HIT: Retorno instantâneo // Removed
	// } // Removed
	// s.mu.RUnlock() // Removed

	// 2. SE NÃO ESTÁ NO CACHE, CHAMAR A IA

	// Timeout agressivo (3s) para proteger o tempo médio de resposta
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	systemPrompt := getPrompt()

	responseFormat := &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONObject,
	}

	resp, err := s.openAIClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: s.modelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("SOLICITAÇÃO: '%s'\n\nRetorne no formato: {\"service_id\": int, \"service_name\": string}", intent),
				},
			},
			ResponseFormat: responseFormat,
		},
	)

	if err != nil {
		return util.FindServiceResponse{Success: false, Error: fmt.Errorf("erro na chamada à API OpenRouter (ou timeout): %w", err).Error()}
	}

	if len(resp.Choices) == 0 {
		return util.FindServiceResponse{Success: false, Error: "a API OpenRouter não retornou resposta (Choices vazio)"}
	}

	// Tentar parsear a resposta JSON da IA
	aiResponseContent := strings.TrimSpace(resp.Choices[0].Message.Content)
	var aiResponse util.AIResponse
	if err := json.Unmarshal([]byte(aiResponseContent), &aiResponse); err != nil {
		fmt.Printf("Erro ao fazer parse do JSON da IA: %v. Conteúdo recebido: %s\n", err, aiResponseContent)
		return util.FindServiceResponse{Success: false, Error: fmt.Errorf("erro ao decodificar a resposta da IA: %w", err).Error()}
	}

	// Validar se o ID retornado existe na nossa lista de serviços válidos
	serviceName, found := util.ValidServices[aiResponse.ServiceID]
	if !found {
		return util.FindServiceResponse{Success: false, Error: fmt.Errorf("o ID de serviço retornado pela IA (%s) é inválido. A IA deve usar apenas IDs válidos", aiResponse.ServiceID).Error()}
	}

	// Cria o dado final, usando o nome oficial do ID para garantir consistência
	finalServiceData := util.ServiceData{
		ServiceID:   aiResponse.ServiceID,
		ServiceName: serviceName,
	}

	// 3. SE SUCESSO, ARMAZENAR NO CACHE (Escrita Protegida)
	return util.FindServiceResponse{Success: true, Data: finalServiceData}
}
