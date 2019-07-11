import os
import sys
import time
  
spkmaster = "spark://hmaster:7077"
work_path = "/usr/local/spark/work"
inbox = "/root/input"
outbox = "/root/output"
deployFile = "deploy.sh"
undeployFile = "undeploy.sh"
master_cfg = "hadoop-master-controller.yaml"
worker_cfg = "hadoop-worker-controller.yaml"
master_svc = "hadoop-master-service.yaml"
nmspace_hp = "namespace-hadoop.yaml"
nmspace = "hadoop"
if len(sys.argv) >= 2:
  nmspace = sys.argv[1]
print(nmspace)

def openFile(fn):
  with open(fn) as f:
    cnt = f.read()
  print(fn, len(cnt))
  return cnt
  
def writeFile(fn, cnt):
  with open(fn, 'w') as f:
    f.write(cnt)
  print(fn, len(cnt))
  
def chnmspace(fn, ns=nmspace):
  with open(fn) as f:
    ls = f.readlines()    
  for i in range(len(ls)):
    if ls[i].find("namespace") != -1:
      ls[i] = ls[i].split(":")[0] + ": " + ns + "\n"    
  with open(fn, 'w') as f:
    f.writelines(ls)
    
def chnmsp_hp(nmsp_hp):
  with open(nmsp_hp) as f:
    cnt = f.readlines()
  tf = False
  for i in range(len(cnt)):
    if cnt[i].find("metadata") != -1:
      tf = True
    elif cnt[i].find("name") != -1 and tf == True and cnt[i].split(":")[0].count(" ") % 2 == 0:
      cnt[i] = cnt[i].split(":")[0] + ": \"" + nmspace + "\"" + "\n"    
    else:
      pass
  with open(nmsp_hp, 'w') as f:
    f.writelines(cnt)   

def chmst_svc(mst_svc):
  with open(mst_svc) as f:
    cnt = f.readlines()
  tf = False
  for i in range(len(cnt)):
    if cnt[i].find("ports") != -1:
      tf = True
    elif cnt[i].find("nodePort") != -1 and tf == True and cnt[i].split(":")[0].count(" ") % 2 == 0:
      ac = cnt[i].split(":")
      cnt[i] = ac[0] + ": " + str(int(ac[1].strip())+10) + "\n"    
    else:
      pass
  with open(mst_svc, 'w') as f:
    f.writelines(cnt)   
    
chnmspace(master_cfg)
chnmspace(worker_cfg)
chnmspace(master_svc)
#if nmspace != "hadoop":
chmst_svc(master_svc)
chnmsp_hp(nmspace_hp)
    
print(openFile(master_cfg))    
print(openFile(master_svc))
print(openFile(worker_cfg))
print(openFile(nmspace_hp)) 

#time.sleep(200) 
os.system("./" + deployFile)   
time.sleep(60) 

def getMaster():
  pods = os.popen("kubectl get pods  --namespace=" + nmspace)
  pods = pods.read()
  nds = [ln.split()[0]  for ln in pods.split("\n") if len(ln)>0 and ln.find("master")!= -1]
  if len(nds) > 0:
    return nds[0]
  else:
    return "" 

hmaster = getMaster()

def getLog(fn):
  pods = os.popen("kubectl get pods  --namespace=" + nmspace)
  pods = pods.read()
  nds = [ln.split()[0]  for ln in pods.split("\n")[1:] if len(ln)>0 and ln.find("master")== -1]
  print(len(nds), nds)
  d={}
  for wk in nds:
    output = os.popen("kubectl exec " + wk + " --namespace=" + nmspace + " -- ls " + work_path)
    output = output.read()
    if len(output) == 0:
      continue
    print(output.split())
    for app in output.split():
      ss = app.split("-")
      d[int(ss[1]+ss[2])] = wk
  print(len(d), d)
  if len(d) == 0:
    return 
  idx = sorted(d)[-1]
  print(idx, d[idx])
  ts = "kubectl exec " + d[idx] + " --namespace=" + nmspace + " -- ls " + work_path + "/app-" + str(idx)[:14] + "-" + str(idx)[14:]
  print("ts", ts)
  tsr = os.popen(ts)
  tsr = tsr.read()
  print(type(tsr), tsr)
  tsr = str(max([int(i) for i in tsr.split()]))
  print(type(tsr), tsr)  
  cpstr = "kubectl cp " + nmspace + "/" + d[idx] + ":" + work_path + "/app-" + str(idx)[:14] + "-" + str(idx)[14:] + "/" + tsr + "/stderr ./" + fn   
  print("cp", cpstr)  
  cp = os.system(cpstr)   

def runSpark(filename, inputMap):  
  executor_cores = inputMap["executor_cores"]
  executor_memory = inputMap["executor_memory"]
  '''worker_number = inputMap["worker_number"]  
  worker_memory = inputMap["worker_memory"]
  worker_cores = inputMap["worker_cores"]  
  cfg = open(worker_cfg)
  ls = cfg.readlines()
  cfg.close()  
  flg = False
  for i in range(len(ls)):
    if ls[i].strip().startswith("replicas:"):      
      ls[i] = ls[i].split(":")[0] + ": " + worker_number + "\n"
    elif ls[i].strip().startswith("limits:"):
      flg = True
    elif ls[i].strip().startswith("cpu:") and flg == True:      
      ls[i] = ls[i].split(":")[0] + ": " + worker_cores + "\n"
    elif ls[i].strip().startswith("memory:") and flg == True:      
      ls[i] = ls[i].split(":")[0] + ": " + worker_memory + "\n"
    elif ls[i].strip().startswith("requests:"):
      flg = False
    else:
      pass
  for ll in ls:
    print(ll)
  cfg1 = open(worker_cfg, 'w')
  cfg1.writelines(ls)
  cfg1.close()
  os.system("kubectl apply -f " + worker_cfg) '''   
  os.system("rm -rf *.result")
  os.system("rm -rf *.log")  
  pfx = "kubectl exec " + hmaster + " --namespace=" + nmspace + " "
  print(pfx)
  isSuccess = 0
  if len(filename) > 0:
    f = filename
    s = f.split(".")[-1]      
    if s == "py":
      print(f)      
      cmd7 = pfx + " -- spark-submit --executor-memory " + executor_memory + " --executor-cores " + executor_cores + " --master " + spkmaster + " " + inbox + "/" + f + " > " + f + ".result"
      print(cmd7)
      isSuccess = os.system(cmd7)      
      if isSuccess == 1:
        isSuccess = False
      else:
        isSuccess = True
      print("write done " + str(isSuccess))
      cmd8 = "kubectl cp " + f + ".result " + nmspace + "/" + hmaster + ":" + outbox
      print(cmd8)      
      os.system(cmd8)      
      getLog(f + ".log")
      os.system("kubectl cp " + f + ".log " + nmspace + "/" + hmaster + ":" + outbox)        
    else:
      pass
  os.system("kubectl exec " + hmaster + " --namespace=" + nmspace + " -- rm -rf " + inbox)
  os.system("kubectl exec " + hmaster + " --namespace=" + nmspace + " mkdir " + inbox)
  result = {}
  result["isSuccess"] = isSuccess
  result["output"] = openFile(filename + ".result")
  result["log"] = openFile(filename + ".log")
  return result
  
if __name__ == "__main__":      
  inputMap={}
  inputMap["executor_cores"] = "1"
  inputMap["executor_memory"] = "512M"
  inputMap["worker_number"] = "1"
  inputMap["worker_memory"] = "512M"
  inputMap["worker_cores"] = '"1"'  
  filename = "spk-example.py"
  print(inputMap, filename)
  result = runSpark(filename, inputMap)  
  print(result["isSuccess"], result["output"], result["log"])
  os.system("./" + undeployFile)