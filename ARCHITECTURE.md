# Arquitetura — FIAP X

## 1. Visão geral de componentes

<svg width="100%" viewBox="0 0 760 620" xmlns="http://www.w3.org/2000/svg" role="img">
  <title>FIAP X — Arquitetura de microsserviços</title>
  <defs>
    <marker id="arr" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
      <path d="M2 1L8 5L2 9" fill="none" stroke="context-stroke" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    </marker>
  </defs>

  <!-- CLIENTE -->
  <rect x="280" y="14" width="200" height="38" rx="8" fill="#F1EFE8" stroke="#5F5E5A" stroke-width="0.8"/>
  <text x="380" y="33" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="500" fill="#444441">Cliente (browser / curl)</text>

  <line x1="380" y1="52" x2="380" y2="86" stroke="#888780" stroke-width="1.5" marker-end="url(#arr)"/>

  <!-- API GATEWAY -->
  <rect x="220" y="88" width="320" height="52" rx="8" fill="#EEEDFЕ" stroke="#534AB7" stroke-width="0.8" fill="#EEEDFE"/>
  <text x="380" y="107" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#3C3489">API Gateway</text>
  <text x="380" y="126" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#534AB7">routing · rate limit · JWT check · injeta X-User-ID</text>

  <!-- seta GW → auth -->
  <path d="M260 140 L260 170 L130 170 L130 196" fill="none" stroke="#73718C" stroke-width="1.5" marker-end="url(#arr)"/>
  <!-- seta GW → upload -->
  <path d="M340 140 L340 196" fill="none" stroke="#73718C" stroke-width="1.5" marker-end="url(#arr)"/>
  <!-- seta GW → status -->
  <path d="M500 140 L500 170 L610 170 L610 196" fill="none" stroke="#73718C" stroke-width="1.5" marker-end="url(#arr)"/>

  <!-- MICROSSERVIÇOS linha 1 -->
  <!-- Auth -->
  <rect x="50" y="198" width="160" height="52" rx="8" fill="#E1F5EE" stroke="#0F6E56" stroke-width="0.8"/>
  <text x="130" y="217" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#085041">Auth Service</text>
  <text x="130" y="236" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#0F6E56">cadastro · login · JWT</text>

  <!-- Upload -->
  <rect x="260" y="198" width="160" height="52" rx="8" fill="#E1F5EE" stroke="#0F6E56" stroke-width="0.8"/>
  <text x="340" y="217" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#085041">Upload Service</text>
  <text x="340" y="236" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#0F6E56">salva vídeo · publica fila</text>

  <!-- Status -->
  <rect x="540" y="198" width="160" height="52" rx="8" fill="#E1F5EE" stroke="#0F6E56" stroke-width="0.8"/>
  <text x="620" y="217" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#085041">Status Service</text>
  <text x="620" y="236" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#0F6E56">lista jobs · download zip</text>

  <!-- seta upload → fila -->
  <line x1="340" y1="250" x2="340" y2="306" stroke="#BA7517" stroke-width="1.5" marker-end="url(#arr)"/>

  <!-- MENSAGERIA container -->
  <rect x="36" y="308" width="688" height="78" rx="10" fill="none" stroke="#EF9F27" stroke-width="0.8" stroke-dasharray="5 3"/>
  <text x="50" y="324" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#854F0B">Mensageria — RabbitMQ</text>

  <!-- fila video.upload -->
  <rect x="56" y="332" width="190" height="40" rx="6" fill="#FAEEDA" stroke="#854F0B" stroke-width="0.8"/>
  <text x="151" y="352" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="500" fill="#633806">video.upload</text>

  <!-- fila notification -->
  <rect x="514" y="332" width="190" height="40" rx="6" fill="#FAEEDA" stroke="#854F0B" stroke-width="0.8"/>
  <text x="609" y="352" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="500" fill="#633806">notification</text>

  <!-- DLQ labels -->
  <text x="151" y="376" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="10" fill="#854F0B">↳ video.upload.dlq (após 3 nacks)</text>
  <text x="609" y="376" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="10" fill="#854F0B">↳ notification.dlq</text>

  <!-- seta fila → worker -->
  <line x1="151" y1="372" x2="151" y2="416" stroke="#BA7517" stroke-width="1.5" marker-end="url(#arr)"/>
  <!-- seta worker → fila notification -->
  <path d="M290 440 L440 440 L440 358 L514 358" fill="none" stroke="#BA7517" stroke-width="1.5" marker-end="url(#arr)"/>
  <!-- seta fila notification → notification service -->
  <line x1="609" y1="372" x2="609" y2="416" stroke="#BA7517" stroke-width="1.5" marker-end="url(#arr)"/>

  <!-- MICROSSERVIÇOS linha 2 -->
  <!-- Worker -->
  <rect x="50" y="418" width="240" height="52" rx="8" fill="#E1F5EE" stroke="#0F6E56" stroke-width="0.8"/>
  <text x="170" y="437" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#085041">Worker Service (N réplicas)</text>
  <text x="170" y="456" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#0F6E56">ffmpeg fps=1 · zip · storage</text>

  <!-- Notification -->
  <rect x="514" y="418" width="190" height="52" rx="8" fill="#E1F5EE" stroke="#0F6E56" stroke-width="0.8"/>
  <text x="609" y="437" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#085041">Notification Service</text>
  <text x="609" y="456" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#0F6E56">e-mail via SMTP</text>

  <!-- PERSISTÊNCIA container -->
  <rect x="36" y="500" width="688" height="102" rx="10" fill="none" stroke="#185FA5" stroke-width="0.8" stroke-dasharray="5 3"/>
  <text x="50" y="516" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#0C447C">Persistência</text>

  <!-- PostgreSQL -->
  <rect x="56" y="524" width="180" height="64" rx="6" fill="#E6F1FB" stroke="#185FA5" stroke-width="0.8"/>
  <text x="146" y="548" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#0C447C">PostgreSQL</text>
  <text x="146" y="568" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#185FA5">users · videos</text>

  <!-- Redis -->
  <rect x="290" y="524" width="180" height="64" rx="6" fill="#E6F1FB" stroke="#185FA5" stroke-width="0.8"/>
  <text x="380" y="548" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#0C447C">Redis</text>
  <text x="380" y="568" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#185FA5">cache de status (TTL 30 s)</text>

  <!-- MinIO -->
  <rect x="524" y="524" width="180" height="64" rx="6" fill="#E6F1FB" stroke="#185FA5" stroke-width="0.8"/>
  <text x="614" y="548" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="13" font-weight="600" fill="#0C447C">MinIO / S3</text>
  <text x="614" y="568" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="11" fill="#185FA5">vídeos · zips</text>

  <!-- setas serviços → persistência -->
  <!-- auth → postgres -->
  <line x1="130" y1="250" x2="130" y2="510" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
  <!-- upload → postgres -->
  <path d="M320 250 L320 490 L146 490 L146 524" fill="none" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
  <!-- upload → minio -->
  <path d="M360 250 L360 490 L614 490 L614 524" fill="none" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
  <!-- status → postgres -->
  <path d="M580 250 L580 490 L170 490" fill="none" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2"/>
  <!-- status → redis -->
  <path d="M640 250 L640 506 L380 506 L380 524" fill="none" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
  <!-- status → minio -->
  <line x1="660" y1="250" x2="660" y2="524" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
  <!-- worker → postgres -->
  <line x1="100" y1="470" x2="100" y2="524" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
  <!-- worker → minio -->
  <path d="M240" y1="470" d="M240 470 L240 494 L590 494 L590 524" fill="none" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
  <!-- notification → postgres -->
  <path d="M560 470 L560 488 L200 488 L200 524" fill="none" stroke="#378ADD" stroke-width="1" stroke-dasharray="3 2" marker-end="url(#arr)"/>
</svg>

---

## 2. Diagrama de sequência — Upload e processamento

<svg width="100%" viewBox="0 0 780 520" xmlns="http://www.w3.org/2000/svg" role="img">
  <title>Sequência de upload e processamento de vídeo</title>
  <defs>
    <marker id="arr2" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
      <path d="M2 1L8 5L2 9" fill="none" stroke="context-stroke" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    </marker>
  </defs>

  <!-- lifeline labels -->
  <text x="40"  y="24" text-anchor="middle" font-family="system-ui,sans-serif" font-size="11" font-weight="600" fill="#444">Usuário</text>
  <text x="140" y="24" text-anchor="middle" font-family="system-ui,sans-serif" font-size="11" font-weight="600" fill="#534AB7">Gateway</text>
  <text x="240" y="24" text-anchor="middle" font-family="system-ui,sans-serif" font-size="11" font-weight="600" fill="#085041">Upload Svc</text>
  <text x="350" y="24" text-anchor="middle" font-family="system-ui,sans-serif" font-size="11" font-weight="600" fill="#085041">Status Svc</text>
  <text x="460" y="24" text-anchor="middle" font-family="system-ui,sans-serif" font-size="11" font-weight="600" fill="#0C447C">PostgreSQL</text>
  <text x="570" y="24" text-anchor="middle" font-family="system-ui,sans-serif" font-size="11" font-weight="600" fill="#0C447C">MinIO</text>
  <text x="670" y="24" text-anchor="middle" font-family="system-ui,sans-serif" font-size="11" font-weight="600" fill="#854F0B">Worker</text>

  <!-- lifelines -->
  <line x1="40"  y1="34" x2="40"  y2="510" stroke="#ccc" stroke-width="1" stroke-dasharray="4 3"/>
  <line x1="140" y1="34" x2="140" y2="510" stroke="#ccc" stroke-width="1" stroke-dasharray="4 3"/>
  <line x1="240" y1="34" x2="240" y2="510" stroke="#ccc" stroke-width="1" stroke-dasharray="4 3"/>
  <line x1="350" y1="34" x2="350" y2="510" stroke="#ccc" stroke-width="1" stroke-dasharray="4 3"/>
  <line x1="460" y1="34" x2="460" y2="510" stroke="#ccc" stroke-width="1" stroke-dasharray="4 3"/>
  <line x1="570" y1="34" x2="570" y2="510" stroke="#ccc" stroke-width="1" stroke-dasharray="4 3"/>
  <line x1="670" y1="34" x2="670" y2="510" stroke="#ccc" stroke-width="1" stroke-dasharray="4 3"/>

  <!-- messages -->
  <!-- 1 POST /videos -->
  <line x1="40" y1="60" x2="132" y2="60" stroke="#534AB7" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="86" y="55" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#534AB7">POST /videos</text>

  <!-- 2 gateway → upload -->
  <line x1="140" y1="80" x2="232" y2="80" stroke="#534AB7" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="190" y="75" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#534AB7">proxy + X-User-ID</text>

  <!-- 3 upload → minio PUT video -->
  <line x1="240" y1="104" x2="562" y2="104" stroke="#0F6E56" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="401" y="99" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#0F6E56">PUT videos/&lt;id&gt;.mp4</text>

  <!-- 4 upload → postgres INSERT -->
  <line x1="240" y1="128" x2="452" y2="128" stroke="#0F6E56" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="346" y="123" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#0F6E56">INSERT video PENDING</text>

  <!-- 5 upload → fila (texto inline) -->
  <line x1="240" y1="152" x2="662" y2="152" stroke="#854F0B" stroke-width="1.5" stroke-dasharray="5 3" marker-end="url(#arr2)"/>
  <text x="451" y="147" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#854F0B">publish video.upload</text>

  <!-- 6 202 Accepted -->
  <line x1="232" y1="176" x2="40" y2="176" stroke="#0F6E56" stroke-width="1.5" stroke-dasharray="4 2" marker-end="url(#arr2)"/>
  <text x="136" y="171" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#0F6E56">202 {video_id}</text>

  <!-- separator -->
  <line x1="20" y1="192" x2="760" y2="192" stroke="#ddd" stroke-width="0.8" stroke-dasharray="2 2"/>
  <text x="390" y="202" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#999">worker consome fila</text>

  <!-- 7 worker UPDATE PROCESSING -->
  <line x1="670" y1="216" x2="468" y2="216" stroke="#854F0B" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="569" y="211" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#854F0B">UPDATE PROCESSING</text>

  <!-- 8 worker GET video -->
  <line x1="670" y1="240" x2="578" y2="240" stroke="#854F0B" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="624" y="235" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#854F0B">GET video</text>

  <!-- 9 ffmpeg -->
  <rect x="646" y="256" width="48" height="20" rx="4" fill="#FAEEDA"/>
  <text x="670" y="266" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="10" fill="#633806">ffmpeg+zip</text>

  <!-- 10 PUT zip -->
  <line x1="670" y1="284" x2="578" y2="284" stroke="#854F0B" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="624" y="279" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#854F0B">PUT zips/&lt;id&gt;.zip</text>

  <!-- 11 UPDATE DONE -->
  <line x1="670" y1="308" x2="468" y2="308" stroke="#854F0B" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="569" y="303" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#854F0B">UPDATE DONE + zip_s3_key</text>

  <!-- 12 ack -->
  <rect x="646" y="320" width="48" height="18" rx="4" fill="#E1F5EE"/>
  <text x="670" y="329" text-anchor="middle" dominant-baseline="central" font-family="system-ui,sans-serif" font-size="10" fill="#0F6E56">ack</text>

  <!-- separator -->
  <line x1="20" y1="348" x2="760" y2="348" stroke="#ddd" stroke-width="0.8" stroke-dasharray="2 2"/>
  <text x="390" y="358" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#999">usuário consulta status</text>

  <!-- 13 GET /videos -->
  <line x1="40" y1="372" x2="132" y2="372" stroke="#534AB7" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="86" y="367" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#534AB7">GET /videos</text>

  <!-- 14 gateway → status -->
  <line x1="140" y1="392" x2="342" y2="392" stroke="#534AB7" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="241" y="387" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#534AB7">proxy + X-User-ID</text>

  <!-- 15 status → redis (cache hit) -->
  <text x="390" y="410" font-family="system-ui,sans-serif" font-size="9" fill="#999">(cache miss → postgres)</text>
  <line x1="350" y1="420" x2="452" y2="420" stroke="#0C447C" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="401" y="415" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#0C447C">SELECT videos</text>

  <!-- 16 200 lista -->
  <line x1="342" y1="444" x2="40" y2="444" stroke="#085041" stroke-width="1.5" stroke-dasharray="4 2" marker-end="url(#arr2)"/>
  <text x="191" y="439" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#085041">200 [{id, status, frame_count}]</text>

  <!-- 17 GET download -->
  <line x1="40" y1="468" x2="132" y2="468" stroke="#534AB7" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="86" y="463" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#534AB7">GET /videos/:id/download</text>

  <!-- 18 status → minio -->
  <line x1="350" y1="488" x2="562" y2="488" stroke="#0F6E56" stroke-width="1.5" marker-end="url(#arr2)"/>
  <text x="456" y="483" text-anchor="middle" font-family="system-ui,sans-serif" font-size="10" fill="#0F6E56">GET zips/&lt;id&gt;.zip → stream</text>
</svg>

---

## 3. Decisões técnicas

### API Gateway sem domínio

O gateway é um **proxy puro** (`net/http/httputil.ReverseProxy`). Não acessa banco,
não conhece usuário nem vídeo. Responsabilidades exclusivas: validar JWT, injetar
`X-User-ID` no header e fazer rate limiting por IP. Cada microsserviço downstream
confia no header sem re-validar o token.

### Separação de responsabilidades por serviço

| Serviço | Escreve em | Lê de |
|---------|-----------|-------|
| Auth | `users` (PostgreSQL) | `users` |
| Upload | `videos` INSERT (PostgreSQL) · MinIO · `video.upload` | — |
| Status | — | `videos` (PostgreSQL + Redis) · MinIO |
| Worker | `videos` UPDATE (PostgreSQL) · MinIO · `notification` | MinIO · `video.upload` |
| Notification | — | `users` (PostgreSQL) |

### Hexagonal / Ports & Adapters

Cada serviço define interfaces de domínio (ports) em `domain/`. Adapters concretos
ficam em subpacotes (`repository/`, `storage/`, `queue/`, `cache/`), substituíveis
sem alterar a lógica de negócio.

```
domain.UserRepository    ←  repository.UserRepo      (pgx)
domain.ObjectStorage     ←  storage.MinIOStorage     (minio-go)
domain.QueuePublisher    ←  queue.RabbitMQPublisher  (amqp091)
domain.StatusCache       ←  cache.RedisCache         (go-redis)
pipeline.VideoProcessor  ←  processor.Processor      (ffmpeg)
```

### Resiliência da fila com Dead-Letter Queue

Mensagens que falham 3 vezes seguidas (`MAX_RETRIES=3`) recebem `nack` sem requeue
e são roteadas automaticamente para a DLQ (`video.upload.dlq`). Isso evita que uma
mensagem corrompida bloqueie workers indefinidamente.

### Cache de status (Redis)

O Status Service implementa **cache-aside**: tenta Redis primeiro (TTL 30 s); em miss
busca no PostgreSQL e repopula o cache. O Worker atualiza o cache após cada `UPDATE`
no banco.

### Escalabilidade horizontal

O Worker não tem estado local — N réplicas consomem a mesma fila em paralelo com
`prefetch=5`. Em Kubernetes, o HPA escala de 2 a 20 réplicas por CPU; para escalar
por profundidade de fila usar KEDA `ScaledObject` com RabbitMQ scaler.

### Notificação desacoplada

O Worker publica em `notification` apenas em caso de erro. O Notification Service
consome essa fila de forma independente e envia e-mail via SMTP. Os dois serviços
evoluem sem acoplamento.

---

## 4. Estrutura de pastas

```
fiapx/
├── api-gateway/              # proxy puro — sem domínio
│   ├── cmd/gateway/main.go
│   ├── config/
│   ├── middleware/           # jwt_check · rate_limit · metrics
│   ├── proxy/                # reverse_proxy.go · router.go
│   └── go.mod
│
├── auth-service/
│   ├── cmd/auth/main.go
│   ├── config/
│   ├── domain/               # User · interfaces
│   ├── handler/              # register · login
│   ├── service/              # AuthService (bcrypt + JWT)
│   ├── repository/           # UserRepo (pgx)
│   ├── mocks/
│   └── go.mod
│
├── upload-service/
│   ├── cmd/upload/main.go
│   ├── config/
│   ├── domain/               # Video · interfaces
│   ├── handler/              # upload
│   ├── service/              # UploadService
│   ├── repository/           # VideoRepo (pgx)
│   ├── storage/              # MinIOStorage
│   ├── queue/                # RabbitMQPublisher
│   ├── mocks/
│   └── go.mod
│
├── status-service/
│   ├── cmd/status/main.go
│   ├── config/
│   ├── domain/
│   ├── handler/              # list · download
│   ├── service/              # StatusService
│   ├── repository/           # VideoRepo (pgx)
│   ├── cache/                # RedisCache
│   ├── storage/              # MinIOStorage
│   ├── mocks/
│   └── go.mod
│
├── worker/
│   ├── cmd/worker/main.go
│   ├── config/
│   ├── consumer/             # AMQP adapter (DLQ · prefetch · ack)
│   ├── pipeline/             # lógica testável
│   ├── processor/            # ffmpeg fps=1 + createZip
│   ├── repository/           # UpdateStatus · UpdateDone (pgx)
│   ├── storage/              # MinIOStorage
│   ├── notification/         # RabbitMQNotifier (publica em fila)
│   ├── mocks/
│   └── go.mod
│
├── notification-service/
│   ├── cmd/notification/main.go
│   ├── config/
│   ├── consumer/             # AMQP consumer
│   ├── mailer/               # SMTPMailer
│   ├── repository/           # UserRepo (busca email)
│   ├── mocks/
│   └── go.mod
│
├── monitoring/
│   ├── prometheus.yml
│   └── grafana/dashboards/
│
├── k8s/                      # Deployment por serviço + Redis
├── terraform/                # AWS EKS
├── legacy/                   # monolito original (referência)
├── scripts/
│   ├── init.sql
│   └── coverage.sh           # gate ≥ 80% por módulo
├── .github/workflows/ci.yml
├── docker-compose.yml
├── sonar-project.properties
└── .env.example
```
