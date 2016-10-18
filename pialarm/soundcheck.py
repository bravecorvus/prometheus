import pygame
import time
from pygame.locals import *

pygame.init()
bleep = pygame.mixer.Sound("police_s.wav")
bleep.play(loops=-1)
print "continuing to do other things"
time.sleep(2.0)
print "still available for more stuff"
