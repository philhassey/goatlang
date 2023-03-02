total = 0
function sphereLoop(o,t)
    o.x = o.x+o.vx*t
    o.y = o.y+o.vy*t
    if o.x < -128 or o.x > 128 then o.vx = -o.vx end
    if o.y < -128 or o.y > 128 then o.vy = -o.vy end
    for _,b in pairs(SPHERES) do
        if b ~= o then
        local dx, dy = b.x-o.x,b.y-o.y
        local dd = (dx*dx+dy*dy)^0.5
        if dd < (o.r+b.r) and dd > 0 then
            dx,dy = dx/dd,dy/dd
            local inc = (o.r+b.r)-dd
            o.x,o.y = o.x-inc*dx,o.y-inc*dy
            b.x,b.y = b.x+inc*dx,b.y+inc*dy
            total = total + 1
        end
    end
    end
end


SPHERES = {}
for i=0,100-1 do
    SPHERES[#SPHERES+1] = {
        x=math.random()*256-128,
        y=math.random()*256-128,
        r=4+math.random()*5,
        vx=math.random()*100-50,
        vy=math.random()*100-50,
        loop=sphereLoop,
    }
end

function tick()
    local t = 1/60.0
    for _,o in pairs(SPHERES) do
        o:loop(t)
    end
end

ts = os.clock()
for i=0,1000-1 do tick() end
print(total)
print(1000 * (os.clock()-ts))