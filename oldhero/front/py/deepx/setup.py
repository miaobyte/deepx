from setuptools import setup, find_packages

setup(
    name='deepx',
    version='0.0.2',
    description='DeepX - 高性能深度学习框架的Python接口',
    author='igor.li',
    author_email='lipeng@mirrorsoft.cn',
    packages=find_packages(),
    install_requires=[
      
    ],
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    url="https://github.com/array2d/deepx",
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires='>=3.7',  # 确保支持数据类型注解
)