package frame

import (
	"time"

	"github.com/palettechain/onRobot/pkg/log"
)

var (
	Tool      = NewPaletteTool()
	startTime = time.Now().Unix()
)

type Method func() bool
type GcFunc func()

type PaletteTool struct {
	//Map name to method
	methodsMap map[string]Method
	//Map method result
	methodsRes map[string]bool
	//gc func
	gc GcFunc
}

func NewPaletteTool() *PaletteTool {
	return &PaletteTool{
		methodsMap: make(map[string]Method, 0),
		methodsRes: make(map[string]bool, 0),
	}
}

func (pt *PaletteTool) RegMethod(name string, method Method) {
	pt.methodsMap[name] = method
}

func (pt *PaletteTool) RegGCFunc(fn GcFunc) {
	pt.gc = fn
}

//Start run
func (pt *PaletteTool) Start(methodsList []string) {
	if len(methodsList) > 0 {
		pt.runMethodList(methodsList)
		return
	}
	log.Info("No method to run")
	return
}

func (pt *PaletteTool) runMethodList(methodsList []string) {
	pt.onStart()
	defer pt.onFinish(methodsList)

	var rest = func(index int) {
		n := len(methodsList)
		if n > 1 && index < n-1 {
			time.Sleep(5 * time.Second)
		}
	}

	for i, method := range methodsList {
		pt.runMethod(i+1, method)
		rest(i)
	}
}

func (pt *PaletteTool) runMethod(index int, methodName string) {
	pt.onBeforeMethodStart(index, methodName)
	method := pt.getMethodByName(methodName)
	if method != nil {
		ok := method()
		pt.onAfterMethodFinish(index, methodName, ok)
		pt.methodsRes[methodName] = ok
		pt.gc()
	}
}

func (pt *PaletteTool) onStart() {
	log.Info("===============================================================")
	log.Info("-------Palette Tool Start-------")
	log.Info("===============================================================")
	log.Info("")
}

func (pt *PaletteTool) onFinish(methodsList []string) {
	failedList := make([]string, 0)
	successList := make([]string, 0)
	for methodName, ok := range pt.methodsRes {
		if ok {
			successList = append(successList, methodName)
		} else {
			failedList = append(failedList, methodName)
		}
	}

	skipList := make([]string, 0)
	for _, method := range methodsList {
		_, ok := pt.methodsRes[method]
		if !ok {
			skipList = append(skipList, method)
		}
	}

	succCount := len(successList)
	failedCount := len(failedList)
	endTime := time.Now().Unix()

	log.Info("===============================================================")
	log.Infof("Palette Tool Finish Total:%v Success:%v Failed:%v Skip:%v, SpendTime:%d sec",
		len(methodsList),
		succCount,
		failedCount,
		len(methodsList)-succCount-failedCount,
		endTime-startTime,
	)

	if succCount > 0 {
		log.Info("---------------------------------------------------------------")
		log.Info("Success list:")
		for i, succ := range successList {
			log.Infof("%d.\t%s", i+1, succ)
		}
	}
	if failedCount > 0 {
		log.Info("---------------------------------------------------------------")
		log.Info("Fail list:")
		for i, fail := range failedList {
			log.Infof("%d.\t%s", i+1, fail)
		}
	}
	if len(skipList) > 0 {
		log.Info("---------------------------------------------------------------")
		log.Info("Skip list:")
		for i, skip := range skipList {
			log.Infof("%d.\t%s", i+1, skip)
		}
	}
	log.Info("===============================================================")
}

func (pt *PaletteTool) onBeforeMethodStart(index int, methodName string) {
	log.Info("===============================================================")
	log.Infof("%d. Start Method:%s", index, methodName)
	log.Info("---------------------------------------------------------------")
}

func (pt *PaletteTool) onAfterMethodFinish(index int, methodName string, res bool) {
	if res {
		log.Infof("Run Method:%s success.", methodName)
	} else {
		log.Infof("Run Method:%s failed.", methodName)
	}
	log.Info("---------------------------------------------------------------")
	log.Info("")
}

func (pt *PaletteTool) getMethodByName(name string) Method {
	return pt.methodsMap[name]
}
