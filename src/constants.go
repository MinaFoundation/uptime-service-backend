package delegation_backend

import "time"

const MAX_SUBMIT_PAYLOAD_SIZE = 1000000
const MAX_BLOCK_SIZE = 50000000
const REQUESTS_PER_PK_HOURLY = 120
const DELEGATION_BACKEND_LISTEN_TO = ":8080"
const TIME_DIFF_DELTA time.Duration = -5*60*1000000000 // -5m

const PK_LENGTH = 35 // why not 33
const SIG_LENGTH = 65 // why not 64

const BASE58CHECK_VERSION_BLOCK_HASH byte = 0x10
const BASE58CHECK_VERSION_PK byte = 0xCB
const BASE58CHECK_VERSION_SIG byte = 0x9A
