【文件夹内容】
1. InvertedIndex.java      实验二主程序（覆盖任务一、任务二）
2. build.sh                编译并打包 jar
3. run_task1.sh            运行任务一（不去停用词）
4. run_task2.sh            运行任务二（去停用词）

【实验要求对应关系】
- 任务一：统计每个词语的总词频和分文档词频，按总词频降序输出。
- 任务二：在任务一基础上读取 hit_stopwords.txt，去除停用词后再统计并排序输出。
- 输出格式：
  词语<TAB>总词频<TAB>文档1:词频<TAB>文档2:词频...

【编译】
chmod +x build.sh
./build.sh

编译成功后会生成：
inverted-index.jar

【运行任务一】
chmod +x run_task1.sh
./run_task1.sh 你的平台用户名

例如：
./run_task1.sh normal4

默认输出目录：
/user/你的平台用户名/BG_Exp2/task1_out

【运行任务二】
chmod +x run_task2.sh
./run_task2.sh 你的平台用户名

例如：
./run_task2.sh normal4

默认输出目录：
/user/你的平台用户名/BG_Exp2/task2_out

【手动运行命令】
hadoop jar inverted-index.jar InvertedIndex '/user/root/Exp2/task1&2' /user/你的平台用户名/BG_Exp2/task1_out
hadoop jar inverted-index.jar InvertedIndex '/user/root/Exp2/task1&2' /user/你的平台用户名/BG_Exp2/task2_out /user/root/Exp2/hit_stopwords.txt

【查看结果】
hdfs dfs -ls /user/你的平台用户名/BG_Exp2/task1_out
hdfs dfs -cat /user/你的平台用户名/BG_Exp2/task1_out/part-r-00000 | head -20

hdfs dfs -ls /user/你的平台用户名/BG_Exp2/task2_out
hdfs dfs -cat /user/你的平台用户名/BG_Exp2/task2_out/part-r-00000 | head -20

【注意】
1. 输入目录 task1&2 中有字符 &，命令里必须加引号，或者写成 task1\&2。
2. 程序会自动执行三个 MapReduce 阶段：
   - 第一阶段：统计 <词语, 文档> 词频
   - 第二阶段：合并成倒排索引行
   - 第三阶段：按总词频降序排序
3. 最终输出目录中 part-r-00000 为全局有序结果。
4. 本文件夹不包含实验报告，按你的要求仅提供代码与运行脚本。
