
from setuptools import setup, find_packages


setup(name='linksim',
    version='0.4.2',
    description='None',
    url='https://github.com/xxx',
    author='auth',
    author_email='xxx@gmail.com',
    license='MIT',
    include_package_data=True,
    zip_safe=False,
    packages=find_packages(),
    install_requires=['requests','pyquery','xlutils', 'termcolor','python-docx', 'pillow', 'xlrd'],
    entry_points={
        'console_scripts': ['linksim=linksim_src.cmd:main', 'office=files_src.cmd:main']
    },

)
