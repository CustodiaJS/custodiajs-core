package vm

import (
	"fmt"
	"sync"
	"time"

	"github.com/CustodiaJS/custodiajs-core/core/consolecache"
	"github.com/CustodiaJS/custodiajs-core/kernel"
	"github.com/CustodiaJS/custodiajs-core/static"
	"github.com/CustodiaJS/custodiajs-core/static/errormsgs"
	"github.com/CustodiaJS/custodiajs-core/types"
	"github.com/CustodiaJS/custodiajs-core/utils"
	"github.com/CustodiaJS/custodiajs-core/vmimage"
)

func (o *CoreVM) GetVMName() string {
	return o.vmImage.GetManifest().Name
}

func (o *CoreVM) GetFingerprint() types.CoreVMFingerprint {
	return types.CoreVMFingerprint(o.Kernel.GetFingerprint())
}

func (o *CoreVM) GetOwner() string {
	return o.vmImage.GetManifest().Owner
}

func (o *CoreVM) GetRepoURL() string {
	return o.vmImage.GetManifest().RepoURL
}

func (o *CoreVM) _routine(scriptContent []byte) {
	// Log
	o.Kernel.LogPrint("", "VM is running")

	// Der Mutex wird verwendet
	o.objectMutex.Lock()

	// Es wird geptüft ob der Stauts des Objektes STILL_WAIT ist
	if o.vmState != static.Starting {
		// Der Mutext wird freigegeben
		o.objectMutex.Unlock()

		// Rückgabe
		return
	}

	// Die Startzeit wird festgelegt
	o.startTimeUnix = uint64(time.Now().Unix())

	// Der Mutex wird freigegeben
	o.objectMutex.Unlock()

	// Das Script wird ausgeführt
	o.runScript(string(scriptContent))

	// Das Script wird als Running Markiert
	o.objectMutex.Lock()

	// Es wird geprüft wie der Aktuele Status des Scriptes ist
	if o.vmState != static.Starting {
		// Der Mutext wird freigegeben
		o.objectMutex.Unlock()

		// Rückgabe
		return
	}

	// Der Status wird auf Running gesetzt
	o.vmState = static.Running

	// Der Mutext wird freigegeben
	o.objectMutex.Unlock()

	// Log
	o.LogPrint("", "Eventloop started")

	// Die Schleife wird solange ausgeführt, solange der Status, running ist.
	// Die Schleife für den Eventloop des Kernels auf
	for o.eventloopForRunner() {
		if err := o.Kernel.ServeEventLoop(); err != nil {
			panic(err)
		}
	}

	// Der Status wird geupdated
	o.objectMutex.Lock()
	o.vmState = static.Closed
	o.objectMutex.Unlock()

	// Log
	o.LogPrint("", "Eventloop stoped")
}

func (o *CoreVM) Serve(syncWaitGroup *sync.WaitGroup) error {
	// Es wird geprüft ob der Server bereits gestartet wurde
	if o.GetState() != static.StillWait && o.GetState() != static.Closed {
		return fmt.Errorf("serveGorutine: vm always running")
	}

	// Die VM wird als Startend Markiert
	o.vmState = static.Starting

	// Diese Funktion wird als Goroutine ausgeführt
	go func() {
		// Die VM wird am leben Erhalten
		o._routine([]byte(o.vmImage.GetMain().Content()))

		// Sollte der Kernel nicht geschlossen sein, wird er beendet
		if !o.Kernel.IsClosed() {
			o.Kernel.Close()
		}

		// Log
		o.Kernel.LogPrint("", "VM is closed")

		// Es wird signalisiert das die VM nicht mehr ausgeführt wird
		syncWaitGroup.Done()
	}()

	// Es ist kein Fehler aufgetreten
	return nil
}

func (o *CoreVM) GetState() types.VmState {
	return o.vmState
}

func (o *CoreVM) GetConsoleOutputWatcher() types.WatcherInterface {
	return o.Kernel.Console().GetOutputStream()
}

func (o *CoreVM) GetStartingTimestamp() uint64 {
	return o.startTimeUnix
}

func (o *CoreVM) runScript(script string) error {
	// Es wird geprüft ob das Script beretis geladen wurden
	if o.scriptLoaded {
		return fmt.Errorf("LoadScript: always script loaded")
	}

	// Es wird markiert dass das Script geladen wurde
	o.scriptLoaded = true

	// Das Script wird ausgeführt
	_, err := o.Kernel.RunScript(script, "main.js")
	if err != nil {
		panic(err)
	}

	// Es ist kein Fehler aufgetreten
	return nil
}

func (o *CoreVM) GetAllSharedFunctions() []types.SharedFunctionInterface {
	extracted := make([]types.SharedFunctionInterface, 0)
	table := o.GloablRegisterRead("rpc")
	if table == nil {
		return extracted
	}

	ctable, istable := table.(map[string]types.SharedFunctionInterface)
	if !istable {
		return extracted
	}

	for _, item := range ctable {
		extracted = append(extracted, item)
	}

	return extracted
}

func (o *CoreVM) GetSharedFunctionBySignature(sourceType types.RPCCallSource, funcSignature *types.FunctionSignature) (types.SharedFunctionInterface, bool, *types.SpecificError) {
	// Es wird versucht die RPC Tabelle zu lesen
	var table interface{}
	if sourceType == static.LOCAL {
		table = o.GloablRegisterRead("rpc")
	} else {
		table = o.GloablRegisterRead("rpc_public")
	}

	// Es wird ermittelt ob die Tabelle gefunden wurde
	if table == nil {
		return nil, false, errormsgs.VM_GET_FUNCTION_BY_SIGNATURE_TABLE_NULL_ERROR("GetSharedFunctionBySignature")
	}

	// Es wird versucht die Tabelle richtig einzulesen
	ctable, istable := table.(map[string]types.SharedFunctionInterface)
	if !istable {
		return nil, false, errormsgs.VM_GET_FUNCTION_RPC_REIGSTER_ERROR("GetSharedFunctionBySignature")
	}

	// Es wird geprüft ob in der Tabelle eine Eintrag für die Funktion vorhanden ist
	result, fResult := ctable[utils.FunctionOnlySignatureString(funcSignature)]
	if !fResult {
		return nil, false, nil
	}

	// Das Ergebniss wird zurückgegeben
	return result, true, nil
}

func (o *CoreVM) hasCloseSignal() bool {
	o.objectMutex.Lock()
	v := bool(o._signal_CLOSE)
	o.objectMutex.Unlock()
	return v
}

func (o *CoreVM) SignalShutdown() {
	// Der Mutex wird angewendet
	o.objectMutex.Lock()

	// Es wird geprüft ob bereits ein Shutdown durchgeführt wurde
	if o._signal_CLOSE || o.vmState == static.Closed {
		o.objectMutex.Unlock()
		return
	}

	// Log
	o.Kernel.LogPrint("", "Signal shutdown")

	// Es wird Signalisiert das ein Close Signal vorhanden ist
	o._signal_CLOSE = true

	// Der Mutex wird freigegeben
	o.objectMutex.Unlock()

	// Der Kernel wird beendet
	o.Kernel.Close()
}

func (o *CoreVM) eventloopForRunner() bool {
	return !o.hasCloseSignal() && !o.Kernel.IsClosed()
}

func (o *CoreVM) IsAllowedXRequested(xrd *types.XRequestedWithData) bool {
	return false
}

func NewCoreVM(core types.CoreInterface, workingDir string, vmImage *vmimage.VmImage, loggingPath types.LOG_DIR) (*CoreVM, error) {
	// Es wird ein neuer Konsolen Stream erzeugt
	consoleStream, err := consolecache.NewConsoleOutputCache(string(loggingPath))
	if err != nil {
		return nil, fmt.Errorf("CoreVM->newCoreVM: " + err.Error())
	}

	// Die Kernel Configurationen werden bereigestellt
	kernelConfig := &kernel.DEFAULT_CONFIG

	// Es wird ein neuer Kernel erzeugt
	vmKernel, err := kernel.NewKernel(consoleStream, kernelConfig, core)
	if err != nil {
		return nil, fmt.Errorf("newCoreVM: " + err.Error())
	}

	// Das Core Objekt wird erstellt
	coreObject := &CoreVM{
		Kernel:        vmKernel,
		core:          core,
		vmImage:       vmImage,
		objectMutex:   &sync.Mutex{},
		vmState:       static.StillWait,
		_signal_CLOSE: false,
	}

	// Es wird versucht die VM mit dem Kernel zu verlinken
	if err := vmKernel.LinkKernelWithCoreVM(coreObject); err != nil {
		return nil, fmt.Errorf("newCoreVM: " + err.Error())
	}

	// Das Objekt wird zurückgegeben
	return coreObject, nil
}