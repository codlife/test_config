import os
import sys
import time
import uuid
  
spkmaster = "spark://hmaster:7077"
spkmasterUI = "http://39.104.74.209:31005/"
work_path = "/usr/local/spark/work"
inbox = "/root/input"
outbox = "/root/output"
nmspace = "hadoop"
e_cores = "1"
e_memory = "512M"
f_name = "spk-example.py"
app_name = "test"
### Usage: python runSpk.py namespace filename executor_cores executor_memory app_name
### Example: python runSpk.py hadoop spk-example.py 1 512M test
if len(sys.argv) >= 6:
  nmspace = sys.argv[1]
  f_name = sys.argv[2]
  e_cores = sys.argv[3]
  e_memory = sys.argv[4]
  app_name = sys.argv[5]
else:
  print("Usage: python runSpk.py namespace filename executor_cores executor_memory app_name")
  print("Example: python runSpk.py hadoop spk-example.py 1 512M test")
print(nmspace, f_name, e_cores, e_memory, app_name)

def openFile(fn):
  with open(fn) as f:
    cnt = f.read()
  print(fn, len(cnt))
  return cnt
  
def writeFile(fn, cnt):
  with open(fn, 'w') as f:
    f.write(cnt)
  print(fn, len(cnt))

def getMaster():
  pods = os.popen("kubectl get pods  --namespace=" + nmspace)
  pods = pods.read()
  nds = [ln.split()[0]  for ln in pods.split("\n") if len(ln)>0 and ln.find("master")!= -1]
  if len(nds) > 0:
    return nds[0]
  else:
    return "" 

hmaster = getMaster()
print(hmaster)

def getAppId(anm):
  tfn = str(uuid.uuid1())
  #print(tfn)
  os.system("curl " + spkmasterUI + " > " + tfn)  
  f = open(tfn)   
  s = f.readlines()
  f.close()  
  #print(s, len(s))
  for i in range(len(s)):
    if s[i].find("appId") > -1 and s[i+2].find("</td>") > -1 and s[i+3].find("<td>") > -1 and i+4 < len(s):      
      #print(s[i].split('"')[1].split("=")[1], s[i+4])
      if s[i+4].strip() == anm: 
        return s[i].split('"')[1].split("=")[1]
        
def getLog(fn, anm):
  pods = os.popen("kubectl get pods  --namespace=" + nmspace)
  pods = pods.read()
  nds = [ln.split()[0]  for ln in pods.split("\n")[1:] if len(ln)>0 and ln.find("master")== -1]
  appId = getAppId(anm)
  #print(len(nds), nds, appId)
  d={}
  for wk in nds:
    output = os.popen("kubectl exec " + wk + " --namespace=" + nmspace + " -- ls " + work_path)
    output = output.read()
    if len(output) == 0:
      continue
    print(output.split())
    for app in output.split():
      if app == appId:
        ss = app.split("-")
        d[int(ss[1]+ss[2])] = wk
  print(len(d), d)
  if len(d) == 0:
    return 
  
  idx = sorted(d)[-1]
  #print(idx, d[idx])
  ts = "kubectl exec " + d[idx] + " --namespace=" + nmspace + " -- ls " + work_path + "/app-" + str(idx)[:14] + "-" + str(idx)[14:]
  #print("ts", ts)
  tsr = os.popen(ts)
  tsr = tsr.read()
  #print(type(tsr), tsr)
  tsr = str(max([int(i) for i in tsr.split()]))
  #print(type(tsr), tsr)  
  cpstr = "kubectl cp " + nmspace + "/" + d[idx] + ":" + work_path + "/app-" + str(idx)[:14] + "-" + str(idx)[14:] + "/" + tsr + "/stderr ./" + fn   
  print("cp", cpstr)  
  cp = os.system(cpstr)   

def runSpark(filename, inputMap):  
  executor_cores = inputMap["executor_cores"]
  executor_memory = inputMap["executor_memory"]  
  app_name = inputMap["app_name"]
  #os.system("rm -rf *.result")
  #os.system("rm -rf *.log")  
  pfx = "kubectl exec " + hmaster + " --namespace=" + nmspace + " "
  #print(pfx)
  isSuccess = 0
  if len(filename) > 0:
    f = filename
    cmdcp = "kubectl cp " + f + " " + nmspace + "/" + hmaster + ":" + inbox  
    print("wjf"+cmdcp)
    os.system(cmdcp) 
    s = f.split(".")[-1]      
    if s == "py":
      print(f)     
      uq = "-" + str(uuid.uuid1()) 
      print("uuid", uq)      
      cmd7 = pfx + " -- spark-submit --executor-memory " + executor_memory + " --executor-cores " + executor_cores + " --master " + spkmaster + " " + inbox + "/" + f + " > " + f + uq + ".result"
      print(cmd7)
      isSuccess = os.system(cmd7)      
      if isSuccess == 1:
        isSuccess = False
      else:
        isSuccess = True
      print("write done " + str(isSuccess))
      cmd8 = "kubectl cp " + f + uq + ".result " + nmspace + "/" + hmaster + ":" + outbox
      print(cmd8)      
      os.system(cmd8)      
      getLog(f +  uq + ".log", app_name)
      os.system("kubectl cp " + f +  uq + ".log " + nmspace + "/" + hmaster + ":" + outbox)        
    else:
      pass
  #os.system("kubectl exec " + hmaster + " --namespace=" + nmspace + " -- rm -rf " + inbox)
  #os.system("kubectl exec " + hmaster + " --namespace=" + nmspace + " mkdir " + inbox)
  os.system("kubectl exec " + hmaster + " --namespace=" + nmspace + " -- rm -rf " + inbox + "/" + filename)
  result = {}
  result["isSuccess"] = isSuccess
  result["output"] = openFile(filename +  uq + ".result")
  result["log"] = openFile(filename +  uq + ".log")
  return result
  
if __name__ == "__main__":      
  inputMap={}
  inputMap["executor_cores"] = e_cores
  inputMap["executor_memory"] = e_memory  
  inputMap["app_name"] = app_name
  filename = f_name
  print(inputMap, filename)
  result = runSpark(filename, inputMap)  
  print(result["isSuccess"], result["output"], result["log"])  
