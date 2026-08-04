package domain

import "time"

// TipoNudge clasifica los ajustes ambientales según su mecanismo conductual.
type TipoNudge string

const (
	NudgeDefaults     TipoNudge = "defaults"
	NudgeSocialProof  TipoNudge = "social_proof"
	NudgeFraming      TipoNudge = "framing"
	NudgeFriccion     TipoNudge = "friccion"
)

// AmbitoNudge define si el nudge aplica a toda la organización o a un grupo específico.
type AmbitoNudge string

const (
	AmbitoGlobal       AmbitoNudge = "global"
	AmbitoDepartamento AmbitoNudge = "departamento"
	AmbitoIndividual   AmbitoNudge = "individual"
)

// Nudge representa un ajuste ambiental que reduce fricción hacia comportamientos deseables.
// No es una notificación — es diseño del entorno.
type Nudge struct {
	ID          int64       `json:"id"`
	Nombre      string      `json:"nombre"`
	Descripcion string      `json:"descripcion"`
	Tipo        TipoNudge   `json:"tipo"`
	Ambito      AmbitoNudge `json:"ambito"`
	TargetID    int64       `json:"target_id,omitempty"` // ID del departamento o empleado si no es global
	Activo      bool        `json:"activo"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
