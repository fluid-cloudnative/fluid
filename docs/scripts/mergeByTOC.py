#!/usr/bin/env python3
# coding: utf8
#
# Generate all-in-one Markdown file for ``doc-cn``
# Tip: 不支持中文文件名
# readme.md 中的目录引用的md多次（或者md的sub heading)，以第一次出现为主
# 每个版本都会生成一个自己的 PDF

from __future__ import print_function, unicode_literals

import re
import os
import sys

followups = []
in_toc = False
contents = []
lang = sys.argv[1]
# pattern=[]()
hyper_link_pattern = re.compile(r'\[(.*?)\]\((.*?)(#.*?)?\)')
# pattern= -- []()
toc_line_pattern = re.compile(r'([\-\+]+)\s\[(.*?)\]\((.*?)(#.*?)?\)')
# pattern= ! []()
image_link_pattern = re.compile(r'!\[(.*?)\]\((.*?)\)')
# pattern=    -+
level_pattern = re.compile(r'(\s*[\-\+]+)\s')
# match all headings
heading_patthern = re.compile(r'(^#+|\n#+)\s')

entry_file = lang + "TOC.md"

# stage 1, parse toc
with open(entry_file) as fp:
    level = 0
    current_level = ""
    for line in fp:
        if not in_toc and line.startswith("## "):
            in_toc = True
        elif in_toc and line.startswith('## '):
            in_toc = False
            # yes, toc processing done
            # contents.append(line[1:]) # skip 1 level TOC
            break
            ## line.strip避免添加空行
        elif in_toc and not line.startswith('#') and line.strip():
            ## get level from space length
            level_space_str = level_pattern.findall(line)[0][:-1]
            #pingcap 文档两个空格一级缩进
            level = len(level_space_str) // 2 + 1 ## python divide get integer

            matches = toc_line_pattern.findall(line)
            if matches:
                for match in matches:
                    fpath = match[2]
                    if fpath.startswith('http'):
                        ## remove list format character `- `, `+ `
                        followups.append(('TOC', level, line.strip()[2:]))
                    elif fpath.endswith('.md'):
                        # remove first slash from the fpath
                        key = ('FILE', level, fpath)
                        if key not in followups:
                            followups.append(key)
            else:
                name = line.strip().split(None, 1)[-1]
                key = ('TOC', level, name)
                if key not in followups:
                    followups.append(key)

        else:
            pass

    # overview part in README.md
    followups.insert(1, ("RAW", 0, fp.read()))
# stage 2, get file heading
file_link_name = {}
title_pattern = re.compile(r'(^#+)\s.*')
for tp, lv, f in followups:
    if tp != 'FILE':
        continue
    try:
        for line in open(lang + f).readlines():
            if line.startswith("#"):
                tag = line.strip()
                break
    except Exception as e:
        print(e)
        tag = ""
    if tag.startswith('# '):
        tag = tag[2:]
    elif tag.startswith('## '):
        tag = tag[3:]
    file_link_name[f] = tag.lower().replace(' ', '-')

def replace_link_wrap(chapter, name):
    # Note: 仅仅支持 hash 匹配，如果在多个文档中有同名 heading 会碰撞
    # 支持 chapter 文档中的 ./ddd.md, xxx.md, xxx.md#xxx 等
    def replace_link(match):
        full = match.group(0)
        link_name = match.group(1)
        link = match.group(2)
        frag = match.group(3)
        if link.startswith('https'):
            return '[{}]({})'.format(link_name,link)
        if link.endswith('.md') or '.md#' in link:
            if link.startswith('../'):
                link=link[3:]
            if not frag:
                for fpath in file_link_name:
                    if link == fpath:
                        frag = '#' + file_link_name[fpath]
            return '[%s](%s)' % (link_name, frag)
        elif link.endswith('.png') or link.endswith('.jpeg') or link.endswith('.svg') or link.endswith('.gif') or link.endswith('.jpg'):
            # special handing for pic
            img_link = re.sub(r'[\.\/]*media\/', './media/', link, count=0, flags=0)
            return '[%s](%s)' % (link_name, img_link)
        else:
            return full

    return hyper_link_pattern.sub(replace_link, chapter)


def replace_heading_func(diff_level=0):

    def replace_heading(match):
        if diff_level == 0:
            return match.group(0)
        else:
            return '\n' + '#' * (match.group(0).count('#') + diff_level) + ' '


    return replace_heading

def replace_img_link(match):
    full = match.group(0)
    link_name = match.group(1)
    link = match.group(2)

    if link.endswith('.png'):
        fname = os.path.basename(link)
        return '![%s](./media/%s)' % (link_name, fname)

# stage 3, concat files
for type_, level, name in followups:
    if type_ == 'TOC':
        contents.append("\n{} {}\n".format('#' * level, name))
    elif type_ == 'RAW':
        contents.append(name)
    elif type_ == 'FILE':
        igore='api_doc.md'
        if igore in name:
            contents.append('# api reference\n\n')
            continue
        try:
            with open(lang + name) as fp:
                chapter = fp.read()
                chapter = replace_link_wrap(chapter, name)
                # chapter = image_link_pattern.sub(replace_img_link, chapter)

                # fix heading level
                off=heading_patthern.findall(chapter)
                if len(off)!=0:
                    diff_level = level - off[0].count('#')

                #print(name, type_, level, diff_level)
                chapter = heading_patthern.sub(replace_heading_func(diff_level), chapter)
                contents.append(chapter)
                contents.append('') # add an empty line
        except Exception as e:
            print(e)
            print("generate file error: ignore!")

# stage 4, generage final doc.md
target_doc_file = lang + 'doc.md'
with open(target_doc_file, 'w') as fp:
    fp.write('\n'.join(contents))
    contents = []
