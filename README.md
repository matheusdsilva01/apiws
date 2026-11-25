## ApiWS

Este é uma simples API WebSocket, construído com GoLang e sem banco de dados, que permite que clientes se conectem em uma sala e enviem mensagens uns aos outros.

Para rodar o projeto, siga os passos abaixo:
- Você precisa ter o Go instalado na sua máquina. Instale aqui: https://golang.org/
- Então instale as dependências executando o comando abaixo na raiz do projeto:
```bash
    go get .
```

- Finalmente, execute o servidor com:
```bash
    go run .
```

### Criar sala

Envie uma requisição `POST /create-room`. O servidor registra uma sala e responde com status `201 Created` e o ID:

```json
{"id":"<room-id>"}
```

Use esse ID para conectar ao WebSocket em `/ws?id=<room-id>`. Salas sem clientes expiram após 8 minutos; salas com clientes conectados não são removidas pelo TTL. A limpeza em memória roda a cada 30 segundos. O estado das salas é mantido somente na memória do processo.
