from datetime import datetime
from datetime import timedelta
from datetime import date
# datetime.now().time()-datetime.strptime('11:11', '%H:%M').time()
x = datetime.combine(date.min, datetime.strptime('11:11', '%H:%M').time())-datetime.combine(date.min, datetime.now().time())
y = datetime.combine(date.min, datetime.now().time())-datetime.combine(date.min, datetime.strptime('11:11', '%H:%M').time())
print(x.total_seconds())
print(datetime.now().date())
# print(y.total_seconds())