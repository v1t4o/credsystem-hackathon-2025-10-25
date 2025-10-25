C4Context

title Diagrama de Componentes - Serviço de Classificação de Intenções

Component(client, "Cliente URA", "Sistema que consome a API de classificação de intenções")
Component(http_server, "Servidor HTTP", "Handler de requisições de API (Go)")
Component(finder_service, "Finder Service", "Lógica de negócio e cache (Go)")
Component(cache, "Cache em Memória", "Armazena intenções já classificadas (Go Map)")
Component(openai_client, "Cliente OpenAI/OpenRouter", "Comunicação com API externa de IA (Go OpenAI SDK)")
Component(ai_model, "Modelo de IA", "Modelo de Classificação de Intenções (OpenAI/OpenRouter)")
Component(prompt_data, "Dados do Prompt", "Template de prompt para a IA (Go)")
Component(types_util, "Tipos e Utilitários", "Estruturas de dados e serviços válidos (Go)")

Rel(client, http_server, "Faz requisições API", "HTTPS")
Rel(http_server, finder_service, "Chama para classificar intenção")
Rel(finder_service, cache, "Lê/Escreve intenção classificada")
Rel(finder_service, openai_client, "Chama API de Chat Completion")
Rel(openai_client, ai_model, "Usa para classificar intenção", "HTTPS")
Rel(finder_service, prompt_data, "Lê template de prompt")
Rel(finder_service, types_util, "Usa estruturas de dados e validações")

UpdateElementStyle(cache, $fontColor="#ffffff", $bgColor="#438DD5", $borderColor="#438DD5")
UpdateElementStyle(finder_service, $fontColor="#ffffff", $bgColor="#2A4D8F", $borderColor="#2A4D8F")
UpdateElementStyle(http_server, $fontColor="#ffffff", $bgColor="#85BB5C", $borderColor="#85BB5C")
