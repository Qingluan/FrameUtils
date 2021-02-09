
import argparse
from files_src.template import defualt_template 
from termcolor import  colored

parser = argparse.ArgumentParser(usage="generate a office html template for flask or tornado ")
parser.add_argument("-g","--generate", default=False, action='store_true', help="set generate true")



def main():
    args = parser.parse_args()
    if args.generate:
        print(defualt_template)
if __name__ == "__main__":
    main()
