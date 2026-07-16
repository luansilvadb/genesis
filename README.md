# DIVI.

**DIVI** é um aplicativo de divisão de despesas domésticas — um _Splitwise_ com alma de livro infantil. Cada casa é um _tenant_ independente onde moradores compartilham cartões de crédito, registram gastos e liquidam saldos por compensação (netting), tudo embrulhado em uma identidade visual lúdica com mascotes animados sobre papel creme.

---

## Funcionalidades

- **Multi-inquilino** — várias casas no mesmo app, cada uma com seus membros, cartões e histórico isolados.
- **Cartões compartilhados** — moradores vinculam cartões de crédito reais; faturas e gastos são consolidados automaticamente.
- **Registro de gastos** — wizard passo a passo com rateio flexível (por valor, porcentagem ou membros selecionados).
- **Netting inteligente** — ao invés de dezenas de transferências, o app calcula a menor quantidade possível de acertos entre os membros.
- **Sincronização em tempo real** — WebSocket mantém todos os dispositivos e abas consistentes, sem refresh.
- **Autenticação JWT + RBAC** — login com e-mail/senha ou Google OAuth; papéis ADMIN e MORADOR com permissões distintas.
- **Mobile-first** — interface pensada para celular, com BottomSheets, Floating Island (barra de navegação flutuante) e modo foco (Zen Mode).

---

## Stack

| Camada      | Tecnologia                                           |
| ----------- | ---------------------------------------------------- |
| **Frontend** | Vue 3 · TypeScript · Tailwind CSS v4 · Vite         |
| **Backend**  | Go 1.23 · Gin · GORM · PostgreSQL · Gorilla WebSocket |
| **Packages** | pnpm v11 · Node ≥24                                  |
| **Testes**   | Vitest (frontend) · `go test` (backend)              |
| **Lint**     | ESLint (frontend) · golangci-lint (backend)          |

---

## Rodando localmente

### Pré-requisitos

- **Node.js** ≥ 24
- **pnpm** ≥ 11
- **Go** ≥ 1.23
- **PostgreSQL** (instância local ou remota)

### 1. Clone o repositório

```bash
git clone https://github.com/luansilvadb/financeiro-divi.git
cd financeiro-divi
```

### 2. Frontend

```bash
pnpm install
pnpm dev          # http://localhost:5173
```

### 3. Backend

Crie `backend-go/.env` a partir do modelo abaixo:

```env
DATABASE_URL=postgres://usuario:senha@localhost:5432/divi?sslmode=disable
JWT_SECRET=uma-chave-secreta-de-pelo-menos-32-caracteres
PORT=3000
GOOGLE_CLIENT_ID=seu-client-id-google
SMTP_HOST=smtp.zoho.com
SMTP_PORT=587
SMTP_USERNAME=seu-email@zoho.com
SMTP_PASSWORD=sua-senha
FRONTEND_URL=http://localhost:5173
CORS_ORIGINS=http://localhost:5173
SWAGGER_ENABLED=false
GIN_MODE=debug
```

Depois suba o servidor:

```bash
cd backend-go
go run ./cmd/server/   # API em http://localhost:3000
```

---

## Testes

### Frontend

```bash
pnpm test              # suite completa (Vitest)
pnpm test:watch        # watch mode
npx vitest --run src/path/to/file.test.ts  # arquivo único
```

### Backend

```bash
cd backend-go
go test ./...                                   # todos os pacotes
go test -v -run TestNome ./internal/...         # teste específico
```

---

## Estrutura do projeto

```
financeiro-divi/
├── index.html                  # Entry point HTML
├── package.json                # Dependências e scripts do frontend
├── pnpm-lock.yaml
├── pnpm-workspace.yaml
├── vite.config.ts
├── tsconfig.json
├── AGENTS.md                   # Guia para agentes de IA
├── DESIGN.md                   # Referência do sistema de design (Family)
│
├── src/                        # 🖥️ Frontend Vue 3
│   ├── main.ts                 # Bootstrap da aplicação
│   ├── App.vue                 # Auth, WebSocket, erros globais
│   ├── main.css                # Tokens Tailwind v4 + estilos base
│   ├── router/                 # Hash-based routing + guards
│   ├── models/
│   │   ├── entities/           # Tipos de domínio (Cartao, Gasto, Dinheiro…)
│   │   ├── repositories/       # Interfaces + implementações HTTP (fetch)
│   │   └── services/           # Lógica de negócio (GastoService, NettingService…)
│   ├── viewmodels/             # Composables reativos (ponte services → views)
│   ├── composables/            # Composables genéricos (useToast, useAsync)
│   ├── views/
│   │   ├── components/
│   │   │   ├── ui/             # Design system (Button, Card, BottomSheet…)
│   │   │   ├── wizard/         # Passos do wizard de lançamento
│   │   │   ├── ledger/         # Feed de atividades, painéis de saldo
│   │   │   └── settings/       # Abas de configurações
│   │   └── screens/            # Telas de alto nível (Login, Dashboard, Onboarding…)
│   └── shared/
│       ├── container.ts        # DI manual (singletons)
│       ├── utils/              # formatarMoeda, rateio, logger, meses…
│       └── validation/         # Schemas Zod + testes de contrato
│
├── backend-go/                 # ⚙️ Backend Go
│   ├── cmd/server/main.go      # Entry point: config, DB, rotas, WS
│   ├── internal/
│   │   ├── config/             # Variáveis de ambiente
│   │   ├── database/           # Bootstrap do banco, AutoMigrate
│   │   ├── model/              # Modelos GORM
│   │   ├── repository/         # Implementações GORM + interfaces
│   │   ├── service/            # AuthService, FinanceiroService, EmailService
│   │   ├── handler/            # Handlers Gin (auth, financeiro, health)
│   │   ├── middleware/          # CORS, JWT, Tenant, RBAC, RateLimit, CSRF
│   │   ├── dto/                # DTOs de request/response
│   │   ├── websocket/          # Hub + client handler Gorilla WS
│   │   └── validator/          # Validadores customizados (senha forte etc.)
│   └── migrations/             # Migrations SQL puras
│
└── public/                     # Assets estáticos (favicon, ícones)
```

---

## Sistema de Design — Family

A identidade visual do DIVI evoca um livro de histórias da Pixar em papel creme.

- **Paleta:** neutros quentes (canvas `#fbfaf9`, stone `#f2f0ed`, charcoal `#343433`) + acentos vibrantes (ember `#ff3e00`, meadow `#00a83d`, sky `#0090ff`).
- **Tipografia:** Fraunces 700 para headings e wordmark, Inter para toda a UI.
- **Cards:** sem sombras externas — apenas `shadow-subtle` (borda interna de 1px).
- **Navegação:** "Floating Island" — barra pill-shaped flutuante com glassmorphism.
- **Mascotes:** criaturas blob animadas com traços de stick e cores primárias.
- **Motion:** transições rápidas (0.2–0.3s) com easing `cubic-bezier(0.19, 1, 0.22, 1)`.

Leia o guia completo em [`DESIGN.md`](./DESIGN.md).

---

## Fluxo de dados

```
[Componente Vue] → [ViewModel] → [Service] → [Repository (fetch)] → [Go API /api/]
                                                                          ↓
[Componente Vue] ← [ViewModel]  ← [SocketService] ← [WebSocket /ws] ← [Hub broadcast]
```

Toda mutação no backend dispara `wsHub.Broadcast(tenantID, evento)`. O `SocketService` do frontend recebe, valida com Zod e dispara eventos internos (`gastos_alterados`, `cartoes_alterados`) que os viewmodels escutam para recarregar os dados.

---

## Fluxo de autenticação

1. Login (e-mail/senha ou Google OAuth) → JWT (HS256) salvo em `localStorage`.
2. `GET /api/auth/me` → lista de casas (tenants) do usuário.
3. Usuário seleciona ou cria uma casa → header `X-Tenant-ID` incluído em toda request.
4. Guards de rota: não autenticado → `/login` · sem casa → `/select-tenant` · ok → `/dashboard`.

---

## Convenções

- **Idioma:** strings de UI e comentários em português; identificadores de código em inglês.
- **Dinheiro:** centavos como `int64` no backend, classe `Dinheiro` no frontend. **Nunca `float`.**
- **Validação de API:** toda resposta da API é validada com Zod no frontend; divergência de contrato quebra explicitamente.
- **Testes:** arquivos `_test.go` (backend) e `.test.ts` ao lado do source (frontend).
- **Rotas:** hash-based (`#/login`, `#/dashboard`) com lazy-loading.

---

## Variáveis de ambiente

### Frontend (`.env`)

| Variável              | Descrição                       | Padrão                    |
| --------------------- | ------------------------------- | ------------------------- |
| `VITE_API_URL`        | URL da API Go                   | `http://localhost:3000`   |
| `VITE_GOOGLE_CLIENT_ID` | Client ID do Google OAuth     | —                         |

### Backend (`backend-go/.env`)

| Variável           | Descrição                          | Padrão                    |
| ------------------ | ---------------------------------- | ------------------------- |
| `DATABASE_URL`     | String de conexão PostgreSQL       | —                         |
| `JWT_SECRET`       | Chave de assinatura JWT (HS256)    | —                         |
| `PORT`             | Porta do servidor                  | `3000`                    |
| `GOOGLE_CLIENT_ID` | Client ID do Google OAuth          | —                         |
| `SMTP_HOST`        | Servidor SMTP (Zoho)               | —                         |
| `SMTP_PORT`        | Porta SMTP                         | —                         |
| `SMTP_USERNAME`    | Usuário SMTP                       | —                         |
| `SMTP_PASSWORD`    | Senha SMTP                         | —                         |
| `FRONTEND_URL`     | Origem CORS                        | —                         |
| `CORS_ORIGINS`     | Origens permitidas (CSV)           | `http://localhost:5173`   |
| `SWAGGER_ENABLED`  | Habilita `/swagger/index.html`     | `false`                   |
| `GIN_MODE`         | `debug` ou `release`               | `debug`                   |

---

## Licença

Proprietário. Todos os direitos reservados.
