# Aurora Home System - Frontend

## Tecnologias

- Angular 18
- PrimeNG 17
- Tailwind CSS 3.4
- TypeScript 5.4

## Paleta de Cores

- **Base**: `#e1c78c`
- **Primary**: `#eda011`
- **Secondary**: `#db6516`
- **Dark**: `#7a6949`
- **Light**: `#adad8e`

## Estrutura

```
src/app/
├── core/                    # Serviços, guards e interceptors
│   ├── services/
│   ├── guards/
│   └── interceptors/
├── shared/                  # Componentes compartilhados
│   ├── components/
│   └── ui/
├── features/                # Módulos de funcionalidades
│   ├── auth/
│   └── dashboard/
├── app.routes.ts
└── app.config.ts
```

## Desenvolvimento

```bash
# Instalar dependências
npm install

# Rodar em desenvolvimento
npm start

# Build para produção
npm run build:prod
```

## Rotas

- `/login` - Tela de login
- `/register` - Tela de registro
- `/dashboard` - Dashboard principal (protegido)

## Docker

O frontend roda em um container Nginx que também atua como proxy reverso para os microserviços.

```bash
# Build da imagem
docker build -t aurora-frontend .

# Rodar container
docker run -p 4200:80 aurora-frontend
```
