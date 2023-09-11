FROM node:lts-alpine as builder

RUN apk add --no-cache python3 make g++ libc6-compat

WORKDIR /app

COPY package*.json ./
RUN npm install
RUN npm install -g ts-node typescript

FROM node:lts-alpine

WORKDIR /app

COPY . .

COPY --from=builder /app/node_modules ./node_modules/
COPY --from=builder /usr/local/bin/ts-node /usr/local/bin/ts-node
COPY --from=builder /usr/local/bin/tsc /usr/local/bin/tsc

ENTRYPOINT [ "./entrypoint.sh" ]
CMD [ "test:integration" ]