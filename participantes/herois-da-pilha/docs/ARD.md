# Arquivo de Revisão de Arquitetura (ARD) - Serviço de Classificação de Intenções Credsystem/Golang SP

## 1. Introdução

Este documento descreve a arquitetura do serviço "Herois da Pilha", um microserviço em Go responsável pela classificação de intenções de usuários para a URA da Credsystem. Ele utiliza um modelo de IA externo (OpenAI/OpenRouter) para analisar o texto da intenção e roteá-la para o serviço mais adequado. O serviço também incorpora um mecanismo de cache para otimizar o desempenho e reduzir custos.

## 2. Visão Geral da Arquitetura

O serviço é construído em Go e segue uma arquitetura baseada em microsserviços. Ele expõe uma API HTTP para receber solicitações de classificação de intenções. A comunicação com o modelo de IA é realizada de forma assíncrona, utilizando um pool de workers e um canal para processamento de jobs. Um cache em memória é utilizado para armazenar respostas de intenções previamente classificadas.

### Componentes Principais:

*   **Servidor HTTP (`main.go`, `handler/api.go`):** Gerencia as requisições HTTP, define os endpoints `/api/find-service` e `/api/healthz`, e delega o processamento da lógica de negócio ao `FinderService`.
*   **`FinderService` (`service/finder.go`):** A lógica central do serviço.
    *   **Cache:** Um mapa em memória (`map[string]util.FindServiceResponse`) com um mutex (`sync.RWMutex`) para garantir acesso seguro e concorrente.
    *   **Cliente OpenAI/OpenRouter:** Responsável pela comunicação com a API externa de IA.
    *   **Pool de Workers:** Goroutines (`worker()`) que consomem jobs de um canal (`jobChannel`) e processam as chamadas à IA de forma assíncrona.
*   **Prompt de IA (`data/prompt.go`):** Contém o template do prompt que é enviado ao modelo de IA para instruí-lo sobre a tarefa de classificação de intenções e os serviços válidos.
*   **Tipos e Utilitários (`util/types.go`):** Define as estruturas de dados (requisições, respostas, dados do serviço, resposta da IA) e o mapa de `ValidServices` para validação e mapeamento de IDs de serviço.
*   **Contêinerização (`Dockerfile`, `docker-compose.yml`):** Define o ambiente de execução do serviço utilizando Docker, facilitando o deploy e a escalabilidade.

## 3. Fluxo de Dados e Comunicação

1.  **Requisição HTTP:** O cliente envia uma requisição POST para `/api/find-service` com o corpo `{"intent": "solicitação do usuário"}`.
2.  **`APIHandler`:** Recebe a requisição, decodifica o JSON e passa a `intent` para o `FinderService`.
3.  **`FinderService` - Cache Check:**
    *   Verifica se a `intent` está presente no cache.
    *   Se sim (cache hit), retorna a resposta em cache imediatamente.
    *   Se não (cache miss), cria um `JobRequest` com a `intent` e um canal de resposta (`ResponseChan`) e o envia para o `jobChannel`.
4.  **`Worker` (Goroutine):**
    *   Uma das goroutines do pool lê um `JobRequest` do `jobChannel`.
    *   Constrói a requisição para a API do OpenRouter, utilizando o `IntentClassificationPrompt` e a `intent` do usuário.
    *   Chama a API `CreateChatCompletion` do OpenAI/OpenRouter com um timeout.
    *   Processa a resposta da IA, decodificando o JSON e validando o `service_id` contra `ValidServices`.
    *   Se a resposta for válida, armazena-a no cache.
    *   Envia a resposta (seja de sucesso ou erro) de volta para o `ResponseChan` específico do job.
5.  **`FinderService` - Espera por Resposta:** O `FinderService` aguarda a resposta no `ResponseChan` e a retorna ao `APIHandler`.
6.  **`APIHandler`:** Codifica a resposta do `FinderService` em JSON e a envia de volta ao cliente.

## 4. Considerações Técnicas

### 4.1. Escalabilidade

*   O uso de um pool de workers com goroutines e canais em Go permite o processamento concorrente de múltiplas requisições à IA, melhorando a capacidade de resposta sob carga.
*   O cache em memória reduz a latência e a carga sobre a API externa, permitindo que o serviço atenda a um maior volume de requisições com intenções repetidas.
*   A contêinerização com Docker e Docker Compose facilita a escalabilidade horizontal, permitindo a execução de múltiplas instâncias do serviço.

### 4.2. Tolerância a Falhas

*   Timeouts são configurados para chamadas à API externa, prevenindo que requisições longas ou falhas da IA bloqueiem o serviço.
*   O tratamento de erros explícito na decodificação de JSON e na validação da resposta da IA garante que o serviço se recupere de payloads malformados ou respostas inesperadas.
*   O uso de mutexes para o cache garante a consistência dos dados em um ambiente concorrente.

### 4.3. Segurança

*   A `OPENROUTER_API_KEY` é carregada de variáveis de ambiente, uma boa prática para evitar credenciais hardcoded no código-fonte.
*   O uso de http.Server com timeouts configurados ajuda a mitigar ataques de Slowloris e outros problemas de conexão.

### 4.4. Otimização de Custos

*   O cache é um componente crucial para otimização de custos, pois evita chamadas redundantes à API de IA para intenções já classificadas. Isso é especialmente importante em modelos de IA onde o custo é por token ou por chamada.

## 5. Próximos Passos / Melhorias Potenciais

*   **Persistência do Cache:** O cache atual é em memória e volátil. Considerar a integração com um cache distribuído (ex: Redis) para persistência e compartilhamento entre instâncias.
*   **Métricas e Monitoramento:** Adicionar métricas (ex: Prometheus) para monitorar o uso da API da IA, taxa de cache hit/miss, latência das requisições, etc.
*   **Tratamento de Erros Aprimorado:** Implementar um tratamento de erros mais robusto, com logs estruturados e identificadores de correlação para facilitar a depuração.
*   **Testes de Carga:** Realizar testes de carga para validar a escalabilidade e o desempenho sob diferentes cenários de uso.
*   **Configuração Centralizada:** Utilizar um sistema de configuração centralizado (ex: Consul, etcd) para gerenciar variáveis de ambiente e configurações de runtime.
*   **Observabilidade:** Adicionar tracing distribuído (ex: OpenTelemetry) para rastrear o fluxo completo de uma requisição.
