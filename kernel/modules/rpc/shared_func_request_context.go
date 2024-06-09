// Author: fluffelpuff
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package kmodulerpc

import (
	"fmt"
	"strings"
	"vnh1/types"
	"vnh1/utils"

	v8 "rogchap.com/v8go"
)

// Sendet eine Erfolgreiche Antwort zurück
func (o *SharedFunctionRequestContext) resolveFunctionCallbackV8(info *v8.FunctionCallbackInfo) *v8.Value {
	// Es wird geprüft ob das Objekt zerstört wurde
	if requestContextIsClosedAndDestroyed(o) {
		panic("destroyed object")
	}

	// Es wird geprüft ob die v8.FunctionCallbackInfo "info" NULL ist
	// In diesem fall muss ein Panic ausgelöst werden, da es keine Logik gibt, wann diese Eintreten kann
	if info == nil {
		panic("vm information callback infor null error")
	}

	// Es wird geprüft ob SharedFunctionRequestContext "o" NULL ist
	if !validateSharedFunctionRequestContext(o) {
		panic("SharedFunctionRequestContext 'o' is empty")
	}

	// Es wird geprüft ob der Vorgang bereits beantwortet wurde,
	// wenn ja wird ein Fehler zurückgegeben dass der Vorgang bereits beantwortet wurde
	if o.wasResponsed() {
		utils.V8ContextThrow(info.Context(), "")
		return nil
	}

	// Speichert alle FunktionsStates ab
	resolves := &types.FunctionCallState{Return: make([]*types.FunctionCallReturnData, 0), State: "ok"}

	// Die Argumente werden umgewandelt
	convertedArguments, err := utils.ConvertV8DataToGoData(info.Args())
	if err != nil {
		utils.V8ContextThrow(info.Context(), err.Error())
		return nil
	}

	// Die Einzelnen Parameter werden abgearbeitet
	for _, item := range convertedArguments {
		resolves.Return = append(resolves.Return, (*types.FunctionCallReturnData)(item))
	}

	// Es wird geprüft ob die Verbindung mit der gegenseite getrennt wurde
	if !o._rprequest.HttpRequest.IsConnected.Bool() {
		utils.V8ContextThrow(info.Context(), "")
		return nil
	}

	// Die Antwort wird geschrieben
	writeRequestReturnResponse(o, resolves)

	// Es ist kein Fehler aufgetreten
	return nil
}

// Sendet eine Rejectantwort zurück
func (o *SharedFunctionRequestContext) rejectFunctionCallbackV8(info *v8.FunctionCallbackInfo) *v8.Value {
	// Es wird geprüft ob die v8.FunctionCallbackInfo "info" NULL ist
	// In diesem fall muss ein Panic ausgelöst werden, da es keine Logik gibt, wann diese Eintreten kann
	if info == nil {
		panic("vm information callback infor null error")
	}

	// Es wird geprüft ob SharedFunctionRequestContext "o" NULL ist
	if !validateSharedFunctionRequestContext(o) {
		// Es wird ein Exception zurückgegeben
		utils.V8ContextThrow(info.Context(), "invalid function share")

		// Undefined wird zurückgegeben
		return v8.Undefined(info.Context().Isolate())
	}

	// Es wird geprüft ob der Vorgang bereits beantwortet wurde,
	// wenn ja wird ein Fehler zurückgegeben dass der Vorgang bereits beantwortet wurde
	if o.wasResponsed() {
		utils.V8ContextThrow(info.Context(), "")
		return nil
	}

	// Die Einzelnen Parameter werden abgearbeitet
	extractedStrings, err := utils.ConvertV8ValuesToString(info.Context(), info.Args())
	if err != nil {
		utils.V8ContextThrow(info.Context(), err.Error())
		return nil
	}

	// Der Finale Fehler wird gebaut
	finalErrorStr := ""
	if len(extractedStrings) > 0 {
		finalErrorStr = strings.Join(extractedStrings, " ")
	}

	// Es wird geprüft ob die Verbindung mit der gegenseite getrennt wurde
	if !o._rprequest.HttpRequest.IsConnected.Bool() {
		utils.V8ContextThrow(info.Context(), "")
		return nil
	}

	// Die Antwort wird zurückgesendet
	writeRequestReturnResponse(o, &types.FunctionCallState{Error: finalErrorStr, State: "failed"})

	// Es ist kein Fehler aufgetreten
	return nil
}

// Wird ausgeführt wenn die Funktion zuende aufgerufen wurde
func (o *SharedFunctionRequestContext) functionCallFinal() error {
	// Es wird geprüft ob das Objekt zerstört wurde
	if requestContextIsClosedAndDestroyed(o) {
		panic("destroyed object")
	}

	// Es wird geprüft ob eine Antwort gesendet wurde
	if !o.wasResponsed() {
		// Der Timer zum abbrechen des Vorganges wird gestartet
		o.startTimeoutTimer()
	}

	// Kernel Log
	o.kernel.LogPrint(fmt.Sprintf("RPC(%s)", o._rprequest.ProcessLog.GetID()), "function call finalized")

	// Es wird nichts zurückgegeben
	return nil
}

// Wird ausgeführt wenn ein Throw durch die Funktion ausgelöst wird
func (o *SharedFunctionRequestContext) functionCallException(msg string) error {
	// Es wird geprüft ob das Objekt zerstört wurde
	if requestContextIsClosedAndDestroyed(o) {
		panic("destroyed object")
	}

	// Die Antwort wird zurückgesendet
	o.resolveChan <- &types.FunctionCallState{Error: msg, State: "exception"}

	// Es wird Signalisiert dass eine Antwort gesendet wurde
	o._wasResponded = true

	// Rückgabe
	return nil
}

// Räumt auf und Zerstört das Objekt
func (o *SharedFunctionRequestContext) clearAndDestroy() {
	// Es wird geprüft ob das Objekt zerstört wurde
	if requestContextIsClosedAndDestroyed(o) {
		panic("destroyed object")
	}

	// Kernel Log
	o.kernel.LogPrint(fmt.Sprintf("RPC(%s)", o._rprequest.ProcessLog.GetID()), "request closed")
}

// Proxy Shielded, Set Timeout funktion
func (o *SharedFunctionRequestContext) proxyShield_SetTimeout(info *v8.FunctionCallbackInfo) *v8.Value {
	fmt.Println("D: TIMER")
	return v8.Undefined(info.Context().Isolate())
}

// Proxy Shielded, Set Interval funktion
func (o *SharedFunctionRequestContext) proxyShield_SetInterval(info *v8.FunctionCallbackInfo) *v8.Value {
	fmt.Println("D: TIMER")
	return v8.Undefined(info.Context().Isolate())
}

// Proxy Shielded, Clear Timeout funktion
func (o *SharedFunctionRequestContext) proxyShield_ClearTimeout(info *v8.FunctionCallbackInfo) *v8.Value {
	fmt.Println("D: TIMER")
	return v8.Undefined(info.Context().Isolate())
}

// Proxy Shielded, Clear Interval funktion
func (o *SharedFunctionRequestContext) proxyShield_ClearInterval(info *v8.FunctionCallbackInfo) *v8.Value {
	fmt.Println("D: TIMER")
	return v8.Undefined(info.Context().Isolate())
}

// Signalisiert dass ein neuer Promises erzeugt wurde und gibt die Entsprechenden Funktionen zurück
func (o *SharedFunctionRequestContext) proxyShield_NewPromise(info *v8.FunctionCallbackInfo) *v8.Value {
	v8Object := v8.NewObjectTemplate(info.Context().Isolate())
	v8Object.Set("resolveProxy", v8.NewFunctionTemplate(info.Context().Isolate(), func(info *v8.FunctionCallbackInfo) *v8.Value {
		o.kernel.LogPrint(fmt.Sprintf("RPC(%s)", o._rprequest.ProcessLog.GetID()), "Promise was resolved")
		return nil
	}))
	v8Object.Set("rejectProxy", v8.NewFunctionTemplate(info.Context().Isolate(), func(info *v8.FunctionCallbackInfo) *v8.Value {
		o.kernel.LogPrint(fmt.Sprintf("RPC(%s)", o._rprequest.ProcessLog.GetID()), "Promise was rejected")
		return nil
	}))

	// Log
	o.kernel.LogPrint(fmt.Sprintf("RPC(%s)", o._rprequest.ProcessLog.GetID()), "New Promise registrated")

	// Das Objekt wird in ein Wert umgewandelt
	obj, _ := v8Object.NewInstance(info.Context())

	// Das Objekt wird zurückgegeben
	return obj.Value
}

// Proxy Shielded, Console Log Funktion
func (o *SharedFunctionRequestContext) proxyShield_ConsoleLog(info *v8.FunctionCallbackInfo) *v8.Value {
	// Es werden alle Stringwerte Extrahiert
	extracted, err := utils.ConvertV8ValuesToString(info.Context(), info.Args())
	if err != nil {
		utils.V8ContextThrow(info.Context(), err.Error())
	}

	// Es wird ein String aus der Ausgabe erzeugt
	outputStr := strings.Join(extracted, " ")

	// Die Ausgabe wird an den Console Cache übergeben
	o.kernel.Console().Log(fmt.Sprintf("RPC(%s):-$ %s", strings.ToUpper(o._rprequest.ProcessLog.GetID()), outputStr))

	// Rückgabe ohne Fehler
	return v8.Undefined(info.Context().Isolate())
}

// Proxy Shielded, Console Log Funktion
func (o *SharedFunctionRequestContext) proxyShield_ErrorLog(info *v8.FunctionCallbackInfo) *v8.Value {
	// Es werden alle Stringwerte Extrahiert
	extracted, err := utils.ConvertV8ValuesToString(info.Context(), info.Args())
	if err != nil {
		utils.V8ContextThrow(info.Context(), err.Error())
		return nil
	}

	// Es wird ein String aus der Ausgabe erzeugt
	outputStr := strings.Join(extracted, " ")

	// Die Ausgabe wird an den Console Cache übergeben
	o.kernel.Console().ErrorLog(fmt.Sprintf("RPC(%s):-$ %s", strings.ToUpper(o._rprequest.ProcessLog.GetID()), outputStr))

	// Rückgabe ohne Fehler
	return v8.Undefined(info.Context().Isolate())
}

// Wartet auf eine Antwort
func (o *SharedFunctionRequestContext) waitOfResponse() (*types.FunctionCallState, error) {
	// Es wird ein neuer Response Waiter erzeugt
	responseWaiter, err := newRequestResponseWaiter(o)
	if err != nil {
		return nil, err
	}

	// Es wird auf den Status gewartet
	finalResolve, err := responseWaiter.WaitOfState()
	if err != nil {
		return nil, err
	}

	// Das Ergebniss wird zurückgegeben
	return finalResolve, nil
}

// Gibt an ob eine Antwort verfügbar ist
func (o *SharedFunctionRequestContext) wasResponsed() bool {
	// Der Mutex wird verwendet
	return o._wasResponded
}

// Startet den Timer, welcher den Vorgang nach erreichen des Timeouts, abbricht
func (o *SharedFunctionRequestContext) startTimeoutTimer() {
}

// Erstellt einen neuen SharedFunctionRequestContext
func newSharedFunctionRequestContext(kernel types.KernelInterface, returnDatatype string, rpcRequest *types.RpcRequest) *SharedFunctionRequestContext {
	// Das Rückgabeobjekt wird erstellt
	returnObject := &SharedFunctionRequestContext{
		resolveChan:     make(chan *types.FunctionCallState),
		_returnDataType: returnDatatype,
		_wasResponded:   false,
		_rprequest:      rpcRequest,
		kernel:          kernel,
	}

	// Das Objekt wird zurückgegeben
	return returnObject
}

// Gibt an ob das Objekt zerstört wurde
func requestContextIsClosedAndDestroyed(o *SharedFunctionRequestContext) bool {
	return o._destroyed
}

// Sendet die Antwort zurück und setzt den Vorgang auf erfolgreich
func writeRequestReturnResponse(o *SharedFunctionRequestContext, returnv *types.FunctionCallState) {
	// Die Goroutine wird verwendet um die Änderungen an dem Request durchzuführen
	go func() {
		// Die Antwort wird zurückgesendet
		o.resolveChan <- returnv

		// Es wird Signalisiert dass eine Antwort gesendet wurde
		o._wasResponded = true
	}()
}