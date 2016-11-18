import pygame
pygame.mixer.init()
pygame.mixer.music.load("/home/pi/Projects/AtomicClock/source/samples/panic_cord.wav")
pygame.mixer.music.play()
while pygame.mixer.music.get_busy() == True:
    continue