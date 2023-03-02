import time

width  = 160
height = 50
its    = 8192
escape = 4.0

def inMandelbrot(cx, cy):
    zx,zy = 0.0,0.0
    for i in range(0,its):
        nx, ny = zx * zx - zy*zy + cx, 2 * zx*zy+cy
        if nx*nx + ny*ny > escape:
            return False
        zx,zy = nx,ny
    return True

def main():
    for y in range(0,height):
        i = float(y*2-height*1) / float(height)
        for x in range(0,width):
            r = float(x*2-width*3/2) / float(width)
            if inMandelbrot(r,i):
                print ("*",end="")
            else:
                print (" ",end="")
        print ("")

ts = time.time()
main()
print(1000 * (time.time()-ts))
