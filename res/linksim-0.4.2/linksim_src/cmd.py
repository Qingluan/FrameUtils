
import argparse
from linksim_src import Url
from linksim_src.ele import detect
from termcolor import  colored

parser = argparse.ArgumentParser(usage="Manager project, can create git , sync , encrypt your repo")
parser.add_argument("-u","--url",default="", help="get group")
parser.add_argument("-d","--detect", default=False, action='store_true', help="set detect true")
parser.add_argument("-r","--rank",default=3, type=int, help="set rank")



def main():
    args = parser.parse_args()
    if args.url != "":
        if not args.detect:
            res = Url.Index(args.url, rank=args.rank)
            for g in res:
                print("-" * 20)
                for l in g:
                    print(l, l.title)
                print()
        else:
            for k,v in detect(args.url).items():
                css, e = v
                print(colored("[%s]"%k, "green", attrs=['bold']),"--"*3, colored(css, 'green',attrs=['underline']))
                print(e)

if __name__ == "__main__":
    main()
