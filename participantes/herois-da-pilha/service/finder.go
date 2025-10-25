package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"herois-da-pilha/data"
	"herois-da-pilha/util"

	"github.com/sashabaranov/go-openai"
)

// FinderService é o struct que gerencia a lógica de IA e o cache.
type FinderService struct {
	openAIClient  *openai.Client
	modelName     string
	cache         map[string]util.FindServiceResponse // Adicionado cache
	mu            sync.RWMutex                        // Mutex para proteger o acesso ao cache
	jobChannel    chan string
	resultChannel chan util.FindServiceResponse
	wg            sync.WaitGroup
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
		openAIClient:  openai.NewClientWithConfig(config),
		modelName:     model,
		cache:         make(map[string]util.FindServiceResponse), // Inicializa o cache
		mu:            sync.RWMutex{},                            // Inicializa o mutex
		jobChannel:    make(chan string),
		resultChannel: make(chan util.FindServiceResponse),
	}

	numWorkers := 5 // Número de goroutines para processar chamadas à IA

	for i := 0; i < numWorkers; i++ {
		s.wg.Add(1)
		go s.worker()
	}

	return s
}

func getPrompt() string {
	return data.IntentClassificationPrompt
}

func (s *FinderService) worker() {
	defer s.wg.Done()
	for intent := range s.jobChannel {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Lógica existente de chamada à IA (sem cache, pois já tratamos isso antes)
		// Este bloco será preenchido posteriormente com a lógica real da IA

		// Simulação de chamada à IA com latência
		// time.Sleep(3 * time.Second)

		// Placeholder para a resposta da IA
		// response := util.FindServiceResponse{Success: true, Data: util.ServiceData{ServiceID: 1, ServiceName: "Simulated Service"}}
		// s.resultChannel <- response

		// A lógica de chamada da IA e processamento da resposta precisa ser movida para cá.
		// Por enquanto, vamos manter o esqueleto.

		// Reativar o cache (se necessário, mas já fizemos isso)
		// s.mu.RLock()
		// if data, ok := s.cache[intent]; ok {
		// 	s.mu.RUnlock()
		// 	s.resultChannel <- data
		// 	continue
		// }
		// s.mu.RUnlock()

		// Chamar a IA
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
			s.resultChannel <- util.FindServiceResponse{Success: false, Error: fmt.Errorf("erro na chamada à API OpenRouter (ou timeout): %w", err).Error()}
			continue
		}

		if len(resp.Choices) == 0 {
			s.resultChannel <- util.FindServiceResponse{Success: false, Error: "a API OpenRouter não retornou resposta (Choices vazio)"}
			continue
		}

		aiResponseContent := strings.TrimSpace(resp.Choices[0].Message.Content)
		var aiResponse util.AIResponse
		if err := json.Unmarshal([]byte(aiResponseContent), &aiResponse); err != nil {
			fmt.Printf("Erro ao fazer parse do JSON da IA: %v. Conteúdo recebido: %s\n", err, aiResponseContent)
			s.resultChannel <- util.FindServiceResponse{Success: false, Error: fmt.Errorf("erro ao decodificar a resposta da IA: %w", err).Error()}
			continue
		}

		serviceName, found := util.ValidServices[aiResponse.ServiceID]
		if !found {
			s.resultChannel <- util.FindServiceResponse{Success: false, Error: fmt.Errorf("o ID de serviço retornado pela IA (%s) é inválido. A IA deve usar apenas IDs válidos", aiResponse.ServiceID).Error()}
			continue
		}

		finalServiceData := util.ServiceData{
			ServiceID:   aiResponse.ServiceID,
			ServiceName: serviceName,
		}

		response := util.FindServiceResponse{Success: true, Data: finalServiceData}

		// Armazenar no cache
		s.mu.Lock()
		s.cache[intent] = response
		s.mu.Unlock()

		s.resultChannel <- response
	}
}

// FindService usa o cache ou o modelo de IA para classificar a intenção.
func (s *FinderService) FindService(intent string) util.FindServiceResponse {
	// 1. TENTAR LER DO CACHE (Leitura Rápida)
	s.mu.RLock()
	if data, ok := s.cache[intent]; ok {
		s.mu.RUnlock()
		return data // Cache HIT: Retorno instantâneo
	}
	s.mu.RUnlock()

	// Enviar a intenção para o canal de jobs e esperar pelo resultado
	s.jobChannel <- intent
	response := <-s.resultChannel
	return response
}
