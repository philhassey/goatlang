import random
import time

total = 0
def sphereLoop(o,t):
    global total
    o.x += o.vx*t
    o.y += o.vy*t
    if o.x < -128 or o.x > 128: o.vx = -o.vx
    if o.y < -128 or o.y > 128: o.vy = -o.vy
    for b in SPHERES:
        if b == o: continue
        dx, dy = b.x-o.x,b.y-o.y
        dd = (dx*dx+dy*dy)**0.5
        if dd < (o.r+b.r) and dd > 0:
            dx,dy = dx/dd,dy/dd
            inc = (o.r+b.r)-dd
            o.x,o.y = o.x-inc*dx,o.y-inc*dy
            b.x,b.y = b.x+inc*dx,b.y+inc*dy
            total += 1

SPHERES = []
class C: pass
for i in range(0,100):
    o = C()
    o.x=random.random()*256-128
    o.y=random.random()*256-128
    o.r=4+random.random()*5
    o.vx=random.random()*100-50
    o.vy=random.random()*100-50
    o.loop=sphereLoop
    SPHERES.append(o)

def tick():
    t = 1/60.0
    for o in SPHERES:
        o.loop(o,t)

ts = time.time()
for i in range(0,1000): tick()
print(total)
print(1000 * (time.time()-ts))
