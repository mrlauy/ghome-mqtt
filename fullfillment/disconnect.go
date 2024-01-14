package fullfillment

import log "log/slog"

type DisconnectResponse struct {
}

func (f *Fullfillment) disconnect(requestId string, payload PayloadRequest) DisconnectResponse {
	log.Info("handle disconnect request", "request", requestId, "payload", payload)
	return DisconnectResponse{}
}
