sequenceDiagram
    participant C as Cliente URA
    participant HS as Servidor HTTP
    participant FS as FinderService
    participant CH as Cache em Memória
    participant OC as Cliente OpenAI/OpenRouter
    participant AIM as Modelo de IA

    C->>HS: POST /api/find-service {intent: "..."}
    activate HS

    HS->>FS: FindService(intent)
    activate FS

    FS->>CH: RLock()
    FS->>CH: Buscar intent no cache
    alt Cache Hit
        CH-->>FS: Resposta em cache
        FS->>CH: RUnlock()
        FS-->>HS: Resposta do serviço
    else Cache Miss
        CH-->>FS: Intent não encontrada
        FS->>CH: RUnlock()

        FS->>FS: Criar JobRequest (intent, ResponseChan)
        FS->>FS: Enviar JobRequest para jobChannel
        note right of FS: Processamento assíncrono por worker

        loop Pool de Workers
            participant W as Worker Goroutine
            FS->>W: JobRequest (via jobChannel)
            activate W

            W->>OC: CreateChatCompletion(ctx, req)
            activate OC
            OC->>AIM: Chamada API de IA (HTTPS)
            activate AIM
            AIM-->>OC: Resposta da IA (JSON)
            deactivate AIM
            OC-->>W: Resposta da API
            deactivate OC

            W->>W: Decodificar e validar resposta da IA
            alt Resposta Válida
                W->>CH: Lock()
                W->>CH: Armazenar resposta no cache
                W->>CH: Unlock()
                W->>FS: Enviar resposta para ResponseChan
            else Resposta Inválida/Erro
                W->>FS: Enviar erro para ResponseChan
            end
            deactivate W
        end

        FS->>FS: Esperar resposta em ResponseChan
        FS-->>HS: Resposta do serviço (via ResponseChan)
    end
    deactivate FS

    HS->>C: Resposta HTTP 200 OK (JSON)
    deactivate HS

    C->>HS: GET /api/healthz
    activate HS
    HS->>C: HTTP 200 OK {status: "ok"}
    deactivate HS
