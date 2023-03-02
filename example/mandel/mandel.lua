width  = 160
height = 50
its    = 8192
escape = 4.0

function inMandelbrot(cx, cy) 
	local zx,zy = 0,0
    for i =0,its do
        local nx, ny = zx * zx - zy*zy + cx, 2 * zx*zy+cy
        if nx*nx + ny*ny > escape then 
			return false
        end
        zx,zy = nx,ny
	end
	return true
end

function main()
    for y=0,height-1 do
		local i = (y*2-height*1) / height
		for x=0,width-1 do
			local r = (x*2-width*3/2) / width
			if inMandelbrot(r,i) then
				io.write("*")
			else
				io.write(" ")
            end
		end
		print("")
	end
end

ts = os.clock()
main()
print(1000 * (os.clock()-ts))
