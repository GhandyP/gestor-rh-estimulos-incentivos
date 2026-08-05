package domain

import (
	"strings"
	"testing"
	"time"
)

// Empleado.Validate

func TestEmpleadoValidate(t *testing.T) {
	valid := Empleado{
		Nombre:       "María García",
		Email:        "maria@empresa.com",
		Cargo:        "Senior Developer",
		Departamento: Departamento("Ingeniería"),
		Activo:       true,
	}

	tests := []struct {
		name    string
		mutate  func(*Empleado)
		wantErr string
	}{
		{"valid employee", func(e *Empleado) {}, ""},
		{"empty nombre", func(e *Empleado) { e.Nombre = "  " }, "nombre"},
		{"empty email", func(e *Empleado) { e.Email = "" }, "email"},
		{"email without @", func(e *Empleado) { e.Email = "maria-empresa.com" }, "email"},
		{"empty cargo", func(e *Empleado) { e.Cargo = "" }, "cargo"},
		{"empty departamento", func(e *Empleado) { e.Departamento = Departamento("") }, "departamento"},
		{"zero fecha ingreso allowed", func(e *Empleado) { e.FechaIngreso = time.Time{} }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := valid
			tt.mutate(&e)
			err := e.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error %q does not mention %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// PerfilMAP.Validate

func TestPerfilMAPValidate(t *testing.T) {
	valid := PerfilMAP{
		EmpleadoID:    1,
		Motivacion:    0.5,
		Habilidad:     0.5,
		Prompt:        0.6,
		Sensibilidad:  SensibilidadDesarrollo,
		Confiabilidad: 0.3,
	}

	tests := []struct {
		name    string
		mutate  func(*PerfilMAP)
		wantErr string
	}{
		{"valid profile", func(p *PerfilMAP) {}, ""},
		{"missing empleado id", func(p *PerfilMAP) { p.EmpleadoID = 0 }, "empleado"},
		{"motivacion above 1", func(p *PerfilMAP) { p.Motivacion = 1.01 }, "motivacion"},
		{"motivacion below 0", func(p *PerfilMAP) { p.Motivacion = -0.1 }, "motivacion"},
		{"habilidad above 1", func(p *PerfilMAP) { p.Habilidad = 1.5 }, "habilidad"},
		{"prompt below 0", func(p *PerfilMAP) { p.Prompt = -0.01 }, "prompt"},
		{"confiabilidad above 1", func(p *PerfilMAP) { p.Confiabilidad = 1.2 }, "confiabilidad"},
		{"unknown sensibilidad", func(p *PerfilMAP) { p.Sensibilidad = Sensibilidad("intrinseco") }, "sensibilidad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := valid
			tt.mutate(&p)
			err := p.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error %q does not mention %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// Umbral.Validate

func TestUmbralValidate(t *testing.T) {
	valid := Umbral{
		EmpleadoID:        1,
		UmbralAbsoluto:    0.30,
		UmbralDiferencial: 0.15,
		UltimoEstimulo:    0.0,
	}

	tests := []struct {
		name    string
		mutate  func(*Umbral)
		wantErr string
	}{
		{"valid initial threshold", func(u *Umbral) {}, ""},
		{"missing empleado id", func(u *Umbral) { u.EmpleadoID = 0 }, "empleado"},
		{"umbral absoluto above 1", func(u *Umbral) { u.UmbralAbsoluto = 1.5 }, "umbral_absoluto"},
		{"umbral diferencial below 0", func(u *Umbral) { u.UmbralDiferencial = -0.1 }, "umbral_diferencial"},
		{"ultimo estimulo above 1", func(u *Umbral) { u.UltimoEstimulo = 1.5 }, "ultimo_estimulo"},
		{"cross-field: intensity without date", func(u *Umbral) { u.UltimoEstimulo = 0.4 }, "fecha_ultimo_estimulo"},
		{"cross-field: date without intensity", func(u *Umbral) {
			now := time.Now()
			u.FechaUltimoEstimulo = &now
		}, "ultimo_estimulo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := valid
			tt.mutate(&u)
			err := u.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error %q does not mention %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// Estimulo.Validate

func TestEstimuloValidate(t *testing.T) {
	valid := Estimulo{
		EmpleadoID:          1,
		Tipo:                "recomendacion_personalizada",
		Contenido:           "Oportunidad de desarrollo profesional para María",
		Intensidad:          0.6,
		Canal:               CanalEmail,
		Estado:              EstadoPendiente,
		FechaIdeal:          time.Now().Add(24 * time.Hour),
		OrigenRecomendacion: "Motor MAP",
	}

	tests := []struct {
		name    string
		mutate  func(*Estimulo)
		wantErr string
	}{
		{"valid pending stimulus", func(e *Estimulo) {}, ""},
		{"missing empleado id", func(e *Estimulo) { e.EmpleadoID = 0 }, "empleado"},
		{"empty tipo", func(e *Estimulo) { e.Tipo = "" }, "tipo"},
		{"empty contenido", func(e *Estimulo) { e.Contenido = " " }, "contenido"},
		{"intensidad above 1", func(e *Estimulo) { e.Intensidad = 1.3 }, "intensidad"},
		{"intensidad below 0", func(e *Estimulo) { e.Intensidad = -0.2 }, "intensidad"},
		{"unknown canal", func(e *Estimulo) { e.Canal = CanalEstimulo("whatsapp") }, "canal"},
		{"unknown estado", func(e *Estimulo) { e.Estado = EstadoEstimulo("borrador") }, "estado"},
		{"cross-field: applied without date", func(e *Estimulo) { e.Estado = EstadoAplicado }, "fecha_aplicado"},
		{"cross-field: date while pending", func(e *Estimulo) {
			now := time.Now()
			e.FechaAplicado = &now
		}, "fecha_aplicado"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := valid
			tt.mutate(&e)
			err := e.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() error %q does not mention %q", err.Error(), tt.wantErr)
			}
		})
	}
}
