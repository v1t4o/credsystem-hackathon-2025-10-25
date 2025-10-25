# Documento de Requisitos de Produto (PRD) - Serviço de Classificação de Intenções Credsystem/Golang SP

## 1. Visão Geral do Produto

O serviço "Herois da Pilha" é um componente backend projetado para classificar a intenção de uma solicitação de usuário em texto livre, direcionando-a para o serviço URA (Unidade de Resposta Audível) da Credsystem mais apropriado. O objetivo é melhorar a experiência do cliente, roteando-o de forma eficiente para o atendimento correto, e otimizar as operações da URA, reduzindo o tempo de espera e a necessidade de intervenção humana desnecessária.

## 2. Metas e Objetivos

*   **Classificação Precisa:** Classificar a intenção do usuário com alta acurácia (alvo: >90%) em relação aos serviços predefinidos da Credsystem.
*   **Redução de Transferências:** Diminuir o número de transferências indevidas para atendimento humano devido a classificações incorretas.
*   **Latência Otimizada:** Fornecer uma resposta de classificação em tempo hábil (alvo: <200ms para 95% das requisições).
*   **Eficiência de Custos:** Minimizar os custos operacionais da API de IA através do uso inteligente de cache e um modelo de IA otimizado para custo/benefício.
*   **Escalabilidade:** Suportar um alto volume de requisições, com capacidade de escalar horizontalmente conforme a demanda.

## 3. Público-Alvo

*   **Usuários Finais da URA Credsystem:** Clientes que interagem com a URA e necessitam de roteamento preciso para seus problemas ou dúvidas.
*   **Equipe de Operações da Credsystem:** Responsável por gerenciar o fluxo da URA e monitorar a eficiência do roteamento.
*   **Desenvolvedores:** Equipes que integrarão seus sistemas com o serviço de classificação de intenções.

## 4. Casos de Uso

### 4.1. Classificação de Intenção do Cliente

**Cenário:** Um cliente liga para a URA e expressa uma intenção.

**Fluxo:**
1.  O sistema da URA captura a fala do cliente e a transcreve para texto.
2.  A URA envia o texto da intenção para o serviço "Herois da Pilha".
3.  O serviço classifica a intenção, identificando o `service_id` e `service_name` mais adequados.
4.  O serviço retorna o `service_id` e `service_name` para a URA.
5.  A URA utiliza essa informação para rotear o cliente para o fluxo de atendimento apropriado.

**Exemplos de Intenções e Classificações Esperadas:**
*   "Quero saber o limite do meu cartão" -> `Consulta Limite / Vencimento do cartão / Melhor dia de compra`
*   "Preciso da segunda via do boleto do meu acordo" -> `Segunda via de boleto de acordo`
*   "Meu cartão não chegou" -> `Status de Entrega do Cartão`
*   "Fui roubado, preciso bloquear meu cartão" -> `Perda e roubo`
*   "Quero falar com um atendente" -> `Atendimento humano`

### 4.2. Verificação de Saúde do Serviço

**Cenário:** Um sistema de monitoramento precisa verificar se o serviço está operante.

**Fluxo:**
1.  O sistema de monitoramento envia uma requisição GET para `/api/healthz`.
2.  O serviço retorna `{"status": "ok"}` se estiver funcionando corretamente.

## 5. Requisitos Funcionais

*   O serviço DEVE expor um endpoint HTTP para classificação de intenções.
*   O serviço DEVE aceitar uma string de intenção como entrada.
*   O serviço DEVE retornar um `service_id` e `service_name` com base na intenção fornecida.
*   O serviço DEVE suportar os 16 serviços predefinidos da Credsystem (conforme listado no `data/prompt.go`).
*   O serviço DEVE utilizar um modelo de IA para a classificação de intenções.
*   O serviço DEVE implementar um mecanismo de cache para respostas de intenções para otimização.
*   O serviço DEVE retornar uma resposta JSON consistente, incluindo campos para sucesso, dados do serviço e mensagens de erro (quando aplicável).
*   O serviço DEVE ter um endpoint de health check (`/api/healthz`).

## 6. Requisitos Não Funcionais

*   **Desempenho:** A latência P95 para classificação de intenções DEVE ser inferior a 200ms.
*   **Disponibilidade:** O serviço DEVE ter uma disponibilidade de 99.9%.
*   **Escalabilidade:** O serviço DEVE ser capaz de escalar horizontalmente para lidar com picos de tráfego de requisições.
*   **Segurança:** A chave da API de IA DEVE ser configurada de forma segura (e.g., variáveis de ambiente).
*   **Custo:** O consumo da API de IA DEVE ser otimizado através do cache para minimizar custos.
*   **Manutenibilidade:** O código DEVE ser legível, modular e bem documentado.
*   **Observabilidade:** O serviço DEVE ser facilmente monitorável, com logs claros e métricas relevantes (futura melhoria).

## 7. Interfaces de Usuário e APIs

### 7.1. Endpoint de Classificação de Intenções

*   **Método:** `POST`
*   **URL:** `/api/find-service`
*   **Corpo da Requisição (JSON):**

```json
{
  "intent": "[texto da intenção do usuário]"
}
```

*   **Corpo da Resposta (JSON - Sucesso):**

```json
{
  "success": true,
  "data": {
    "service_id": [ID do serviço (int)],
    "service_name": "[Nome do serviço (string)]"
  },
  "error": ""
}
```

*   **Corpo da Resposta (JSON - Erro):**

```json
{
  "success": false,
  "data": {
    "service_id": 0,
    "service_name": ""
  },
  "error": "[Mensagem de erro]"
}
```

### 7.2. Endpoint de Health Check

*   **Método:** `GET`
*   **URL:** `/api/healthz`
*   **Corpo da Resposta (JSON):**

```json
{
  "status": "ok"
}
```

## 8. Considerações Futuras

*   **Aprimoramento do Modelo:** Possibilidade de treinar ou ajustar o modelo de IA para melhor se adaptar a nuances específicas da Credsystem.
*   **Feedback Loop:** Implementar um mecanismo para coletar feedback sobre a precisão das classificações, permitindo a melhoria contínua do modelo ou prompt.
*   **Internacionalização:** Suporte a múltiplos idiomas, caso a Credsystem expanda suas operações.
