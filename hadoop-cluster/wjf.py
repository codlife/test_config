import os
result = os.popen("cat test.spark")
print(result.read())
