# play sound
import pygame.mixer
from pygame.mixer import Sound

pygame.mixer.init()

drum= Sound("samples/panic_cord.wav")

while True:
    drum.play()