# play sound
import pygame.mixer
from pygame.mixer import Sound

pygame.mixer.init()

panic_cord = Sound("/home/pi/Projects/AtomicClock/source/samples/panic_cord.wav")

while True:
    panic_cord.play()