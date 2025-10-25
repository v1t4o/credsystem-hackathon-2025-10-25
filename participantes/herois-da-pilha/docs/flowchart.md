flowchart TD
    start[Início: Requisição POST /api/find-service] --> A{Decodificar JSON da requisição?}

    A -- Sim --> B{Intent está no Cache?}
    A -- Não / Erro --> F1[Erro: Requisição inválida]

    B -- Sim (Cache Hit) --> C[Retornar resposta do Cache] --> end_success(Fim: Sucesso)
    B -- Não (Cache Miss) --> D[Criar JobRequest e enviar para jobChannel]

    D --> E[Aguardar resposta em ResponseChan]

    subgraph Worker Goroutine (Processamento Assíncrono)
        E_worker[Receber JobRequest do jobChannel] --> F{Construir Requisição OpenAI/OpenRouter}
        F --> G[Chamar API CreateChatCompletion]
        G -- Sucesso --> H{Decodificar e validar Resposta da IA?}
        G -- Erro/Timeout --> I[Erro: Falha na API ou Timeout]

        H -- Sim --> J[Armazenar resposta no Cache]
        J --> K[Enviar resposta para ResponseChan do Job]
        H -- Não --> I

        I --> K
    end

    K --> E

    E --> L{Resposta recebida do ResponseChan?}
    L -- Sim --> end_success
    L -- Não / Timeout --> F2[Erro: Timeout ou falha interna]

    subgraph Health Check
        start_health[Início: Requisição GET /api/healthz] --> H_A[Retornar {status: "ok"}]
        H_A --> end_health(Fim: Health Check Sucesso)
    end

    F1 --> end_error(Fim: Erro)
    F2 --> end_error
