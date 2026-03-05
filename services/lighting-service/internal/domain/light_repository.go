package domain

// LightRepository define o contrato para persistência de luzes
type LightRepository interface {
	FindByID(id string) (*Light, error)
	FindByOwnerID(ownerID string) ([]*Light, error)
	FindAll() ([]*Light, error)
	Save(light *Light) error
}

// DeviceClient define o contrato para comunicação com o ESP32
type DeviceClient interface {
	TurnOn(deviceID string) error
	TurnOff(deviceID string) error
	GetStatus(deviceID string) (LightState, error)
}

// Device representa um dispositivo físico ESP32 registrado
type Device struct {
	ID        string
	IPAddress string
	Port      int
	LastSeen  string
}

// DeviceRegistry define o contrato para registro de dispositivos
type DeviceRegistry interface {
	Register(device *Device) error
	FindByID(id string) (*Device, error)
}
