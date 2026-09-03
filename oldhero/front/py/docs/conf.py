import os
import sys
sys.path.insert(0, os.path.abspath('..'))  # 添加项目根目录到路径

# 扩展配置
extensions = [
    'sphinx.ext.autodoc',     # 自动提取文档字符串
    'sphinx.ext.napoleon',    # 支持 Google 和 NumPy 风格的文档
    'sphinx.ext.mathjax',     # 支持数学公式渲染
    'sphinx.ext.viewcode',    # 链接到源代码
]

# 主题设置
html_theme = 'sphinx_rtd_theme'  # 使用 Read the Docs 主题

# 项目信息
project = 'deepx'
copyright = '2024, Your Name'
author = 'Your Name' 