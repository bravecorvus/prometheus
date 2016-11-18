# play sound
# import pygame.mixer
# from pygame.mixer import Sound

# pygame.mixer.init()

# panic_cord = Sound("/home/pi/Projects/AtomicClock/source/samples/panic_cord.wav")

# while True:
#     panic_cord.play()

import pygame
pygame.mixer.init()
pygame.mixer.music.load("myFile.wav")
pygame.mixer.music.play()
while pygame.mixer.music.get_busy() == True:
    continue