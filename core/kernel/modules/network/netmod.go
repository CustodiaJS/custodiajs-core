package kmodulenet

import (
	"fmt"
	"vnh1/types"

	v8 "rogchap.com/v8go"
)

type ModuleNetwork struct {
	context *v8.Context
}

func (o *ModuleNetwork) Init(kernel types.KernelInterface) error {
	// Das Consolen Objekt wird definiert
	console := v8.NewObjectTemplate(kernel.Isolate())

	// Der Kontext wird abgespeichert
	o.context = v8.NewContext(kernel.Isolate())

	// Das Console Objekt wird final erzeugt
	consoleObj, err := console.NewInstance(o.context)
	if err != nil {
		return fmt.Errorf("Kernel->_new_kernel_load_console_module: " + err.Error())
	}

	// Das Objekt wird als Import Registriert
	if err := kernel.AddImportModule("net", consoleObj.Value); err != nil {
		return fmt.Errorf("ModuleNetwork->Init: " + err.Error())
	}

	// Kein Fehler
	return nil
}

func (o *ModuleNetwork) GetName() string {
	return "console"
}

func NewNetworkModule() *ModuleNetwork {
	return new(ModuleNetwork)
}