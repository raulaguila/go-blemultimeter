# Análise Técnica do Projeto `go-blemultimeter`

Data da análise: 2026-02-15  
Escopo: arquitetura de software, comunicação BLE, parser de protocolos e qualidade geral da base.

## 1) Visão geral da solução

O projeto implementa leitura de dados BLE de multímetros e tradução de payloads proprietários para saída estruturada (`valor`, `unidade`, `flags`).

Componentes principais:
- Entrada CLI em `cmd/go-multimeter/main.go`.
- Orquestração de leitura em `internal/reader/reader.go`.
- Camada BLE em `pkg/bluetooth/bluetooth.go`.
- Decodificadores de protocolo:
  - `pkg/multimeter/owon/ow18e.go`
  - `pkg/multimeter/fs9721/fs9721.go`
- Contrato de extensão em `internal/domain/multimeter.go`.

## 2) Pontos fortes

1. Separação clara de responsabilidades por pacote
- A divisão `cmd` -> `reader` -> `bluetooth` -> `multimeter/*` está simples e direta.
- O contrato `Multimeter` desacopla parser de protocolo da infraestrutura BLE (`internal/domain/multimeter.go:3`).

2. Arquitetura extensível para novos dispositivos
- Novos multímetros podem ser adicionados implementando `ProcessArray(...)` e novos `configs.go`.
- O encapsulamento por vendor/modelo em `pkg/multimeter/*` está adequado.

3. Documentação técnica de protocolo acima da média
- Os READMEs dos protocolos (`pkg/multimeter/owon/README.md` e `pkg/multimeter/fs9721/README.md`) explicam payloads, bits e semântica.
- Isso reduz curva de aprendizado para manutenção e onboarding.

4. Base pequena, legível e com baixo acoplamento estrutural
- O código é relativamente curto e fácil de navegar.
- Dependência BLE única (`tinygo.org/x/bluetooth`) simplifica entendimento do stack.

5. Compilação/checagem básica saudáveis
- `go test ./...` executa sem falhas.
- `go vet ./...` sem achados.

## 3) Pontos de melhoria (priorizados)

### Alta prioridade

1. Tratamento de erro via `panic` reduz robustez operacional
- `pkg/alert/alert.go:3` transforma qualquer erro em `panic`.
- Isso é arriscado para um processo que depende de BLE (ambiente naturalmente instável: timeout, interferência, desconexões).
- Impacto: queda total da aplicação por erro transitório.

2. Risco de `panic` por falta de validação de tamanho no parser OW18E
- `pkg/multimeter/owon/ow18e.go:34-37` acessa índices fixos (`binArray[0]`, `binArray[1]`) sem checar tamanho do frame.
- `pkg/multimeter/owon/ow18e.go:64-72` usa `byteArray[4]` e `byteArray[5]` assumindo 6 bytes.
- Em frame truncado/ruído de notificação, há risco de `index out of range`.

3. Ciclo de vida de goroutines/canais pode vazar em cenários de desconexão
- `internal/reader/reader.go:63-65` faz receive bloqueante (`<-chNotify`) dentro do loop de conexão.
- Se desconectar sem novo pacote, a goroutine pode ficar bloqueada.
- `pkg/bluetooth/bluetooth.go:133-135` envia em canal sem controle de fechamento/cancelamento.

4. Estado concorrente sem sincronização explícita
- Campo `connected` é lido/escrito por múltiplas rotas (`Connected`, `Disconnect`, loops goroutine) em `pkg/bluetooth/bluetooth.go`.
- Sem mutex/atômico/contexto coordenado, há risco de data race em execução concorrente.

### Média prioridade

5. Ausência de suíte de testes automatizados
- Não há arquivos `_test.go` no projeto.
- Falta cobertura para:
  - Parsing de frames válidos/ inválidos.
  - Casos limite (`L`, negativo, mudança de range, unidades compostas).
  - Regressão de protocolos.

6. Experiência de CLI e documentação com inconsistência pontual
- README raiz cita `go run cmd/app/main.go`, mas o entrypoint atual é `cmd/go-multimeter/main.go`.
- Pode gerar fricção para novos usuários.

7. Contrato e nomenclatura com ruído
- Método `ProcessArray` está com grafia não convencional ("ProcessArray").
- `externChanel` também contém typo em `internal/reader/reader.go:56`.
- Não quebra funcionamento, mas reduz clareza e profissionalismo da API.

### Baixa prioridade

8. Melhorias de modelagem de dados
- Uso de `[3]interface{}` em `internal/reader/reader.go:56` perde type-safety.
- Um struct explícito melhoraria legibilidade e evolução.

9. Mapeamentos de unidade/flags podem ganhar estratégia declarativa
- Hoje há blocos hardcoded extensos (`extractUnit`, `extractFlags`), funcionais mas suscetíveis a erro humano em manutenção.

## 4) Análise de arquitetura BLE

Pontos positivos:
- Fluxo BLE direto: `Connect` -> descoberta de characteristic -> notifier/writer.
- Separação entre notificação (leitura) e escrita em `Bluetooth.StartNotifier` e `Bluetooth.StartWriter`.

Riscos arquiteturais atuais:
- Falta de contexto de cancelamento propagado fim a fim (reader + bluetooth + parser).
- Falta de política de reconexão (retry/backoff/jitter).
- Falta de telemetria estruturada de erros BLE (tipos de falha, contagem, última conexão estável).

## 5) Recomendações práticas (roadmap)

### Fase 1 (hardening, curto prazo)
1. Substituir `panic` por erros propagados e mensagens de recuperação.
2. Validar tamanho dos frames antes de indexar payloads em `ow18e`.
3. Introduzir `context.Context` em loops de notificação para encerramento limpo.
4. Adicionar timeout e retry controlado na conexão BLE.

### Fase 2 (qualidade e confiabilidade)
1. Criar testes table-driven para `ow18e` e `fs9721` com fixtures de frame.
2. Incluir testes de casos inválidos/truncados para evitar `panic`.
3. Trocar `[3]interface{}` por struct tipada de leitura.

### Fase 3 (produto e DX)
1. Atualizar README de uso/quickstart com comando correto.
2. Padronizar nomenclaturas (`ProcessArray`, `externalChannel`, etc.).
3. Adicionar modo de saída estruturada (JSON/CSV) para integração com outros sistemas.

## 6) Resumo executivo

A base está bem organizada e tem valor técnico real na engenharia reversa dos protocolos BLE de multímetros, com boa separação por domínios. O principal gap é robustez operacional: hoje erros transitórios e payloads inesperados podem derrubar o processo. O caminho de maior retorno é endurecer tratamento de erro/ciclo de vida concorrente e estabelecer testes de parser com fixtures reais.

Com essas melhorias, o projeto evolui de um ótimo protótipo funcional para uma biblioteca/ferramenta mais confiável para uso contínuo em ambiente real.
