import sys

from PIL import Image

fileToOpen = sys.argv[1]

im = Image.open(fileToOpen)

im.show()

sys.exit(0)
