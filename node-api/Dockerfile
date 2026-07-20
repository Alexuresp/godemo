FROM node:22-alpine AS deps
WORKDIR /src
COPY package.json package-lock.json* ./
RUN npm install --production --frozen-lockfile

FROM node:22-alpine
WORKDIR /app
COPY --from=deps /src/node_modules ./node_modules
COPY src/ ./src/
COPY package.json ./

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001 -G nodejs

USER nodejs

EXPOSE 3000
ENV NODE_ENV=production

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000/healthz', (r) => { process.exit(r.statusCode === 200 ? 0 : 1) })"

CMD ["node", "src/index.js"]
