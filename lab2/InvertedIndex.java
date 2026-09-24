import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.URI;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.TreeMap;
import java.util.HashSet;
import java.util.Set;

import org.apache.hadoop.conf.Configuration;
import org.apache.hadoop.conf.Configured;
import org.apache.hadoop.fs.FileSystem;
import org.apache.hadoop.fs.Path;
import org.apache.hadoop.io.IntWritable;
import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.io.WritableComparable;
import org.apache.hadoop.io.WritableComparator;
import org.apache.hadoop.mapreduce.Job;
import org.apache.hadoop.mapreduce.Mapper;
import org.apache.hadoop.mapreduce.Reducer;
import org.apache.hadoop.mapreduce.lib.input.FileInputFormat;
import org.apache.hadoop.mapreduce.lib.input.FileSplit;
import org.apache.hadoop.mapreduce.lib.input.TextInputFormat;
import org.apache.hadoop.mapreduce.lib.output.FileOutputFormat;
import org.apache.hadoop.mapreduce.lib.output.TextOutputFormat;
import org.apache.hadoop.util.Tool;
import org.apache.hadoop.util.ToolRunner;

/**
 * 实验二：文档倒排索引
 *
 * 功能：
 * 1. 任务一：统计每个词语的总词频和分文档词频，并按总词频降序输出。
 * 2. 任务二：在任务一基础上读取停用词表，过滤停用词后再统计并排序输出。
 *
 * 运行方式：
 *   hadoop jar inverted-index.jar InvertedIndex <input> <output>
 *   hadoop jar inverted-index.jar InvertedIndex <input> <output> <stopwords>
 *
 * 例如：
 *   hadoop jar inverted-index.jar InvertedIndex '/user/root/Exp2/task1&2' /user/normal4/BG_Exp2/task1_out
 *   hadoop jar inverted-index.jar InvertedIndex '/user/root/Exp2/task1&2' /user/normal4/BG_Exp2/task2_out /user/root/Exp2/hit_stopwords.txt
 */
public class InvertedIndex extends Configured implements Tool {

    private static final String SEP = "\u0001";
    private static final IntWritable ONE = new IntWritable(1);

    /**
     * 第一阶段 Mapper：
     * 输入一行评论，输出 <词语+分隔符+文档名, 1>
     */
    public static class WordDocMapper extends Mapper<LongWritable, Text, Text, IntWritable> {
        private final Text outKey = new Text();
        private final Set<String> stopWords = new HashSet<String>();
        private boolean useStopWords;
        private String docName;

        @Override
        protected void setup(Context context) throws IOException, InterruptedException {
            useStopWords = context.getConfiguration().getBoolean("invertedindex.use.stopwords", false);

            FileSplit split = (FileSplit) context.getInputSplit();
            docName = stripExtension(split.getPath().getName());

            if (useStopWords) {
                File localStopwords = new File("hit_stopwords.txt");
                if (!localStopwords.exists()) {
                    throw new IOException("未找到本地缓存的停用词文件 hit_stopwords.txt");
                }
                loadStopWords(localStopwords);
            }
        }

        private void loadStopWords(File file) throws IOException {
            try (BufferedReader br = new BufferedReader(
                    new InputStreamReader(new FileInputStream(file), StandardCharsets.UTF_8))) {
                String line;
                while ((line = br.readLine()) != null) {
                    line = line.trim();
                    if (!line.isEmpty()) {
                        stopWords.add(line);
                    }
                }
            }
        }

        @Override
        protected void map(LongWritable key, Text value, Context context) throws IOException, InterruptedException {
            String line = value.toString();
            if (line == null || line.isEmpty()) {
                return;
            }

            int commaPos = line.indexOf(',');
            if (commaPos < 0 || commaPos == line.length() - 1) {
                return;
            }

            String content = line.substring(commaPos + 1).trim();
            if (content.isEmpty()) {
                return;
            }

            String[] words = content.split("\\s+");
            for (String word : words) {
                if (word == null) {
                    continue;
                }
                word = word.trim();
                if (word.isEmpty()) {
                    continue;
                }
                if (useStopWords && stopWords.contains(word)) {
                    continue;
                }
                outKey.set(word + SEP + docName);
                context.write(outKey, ONE);
            }
        }
    }

    /** 第一阶段 Combiner/Reducer：统计同一 <词语, 文档> 的词频 */
    public static class SumReducer extends Reducer<Text, IntWritable, Text, IntWritable> {
        private final IntWritable outValue = new IntWritable();

        @Override
        protected void reduce(Text key, Iterable<IntWritable> values, Context context)
                throws IOException, InterruptedException {
            int sum = 0;
            for (IntWritable value : values) {
                sum += value.get();
            }
            outValue.set(sum);
            context.write(key, outValue);
        }
    }

    /**
     * 第二阶段 Mapper：
     * 把第一阶段输出的 <词语+分隔符+文档, 词频>
     * 转成 <词语, 文档+分隔符+词频>
     */
    public static class PostingListMapper extends Mapper<LongWritable, Text, Text, Text> {
        private final Text outKey = new Text();
        private final Text outValue = new Text();

        @Override
        protected void map(LongWritable key, Text value, Context context) throws IOException, InterruptedException {
            String line = value.toString();
            if (line == null || line.isEmpty()) {
                return;
            }

            int tabPos = line.indexOf('\t');
            if (tabPos < 0 || tabPos == line.length() - 1) {
                return;
            }

            String compositeKey = line.substring(0, tabPos);
            String countStr = line.substring(tabPos + 1).trim();

            int sepPos = compositeKey.lastIndexOf(SEP);
            if (sepPos < 0 || sepPos == compositeKey.length() - 1) {
                return;
            }

            String word = compositeKey.substring(0, sepPos);
            String docName = compositeKey.substring(sepPos + 1);

            outKey.set(word);
            outValue.set(docName + SEP + countStr);
            context.write(outKey, outValue);
        }
    }

    /**
     * 第二阶段 Reducer：
     * 生成倒排索引行：<词语, 总词频\t文档1:词频\t文档2:词频...>
     */
    public static class PostingListReducer extends Reducer<Text, Text, Text, Text> {
        private final Text outValue = new Text();

        @Override
        protected void reduce(Text key, Iterable<Text> values, Context context)
                throws IOException, InterruptedException {
            TreeMap<String, Integer> docCountMap = new TreeMap<String, Integer>();
            int total = 0;

            for (Text value : values) {
                String s = value.toString();
                int sepPos = s.lastIndexOf(SEP);
                if (sepPos < 0 || sepPos == s.length() - 1) {
                    continue;
                }
                String docName = s.substring(0, sepPos);
                int count = Integer.parseInt(s.substring(sepPos + 1));
                total += count;

                Integer old = docCountMap.get(docName);
                if (old == null) {
                    docCountMap.put(docName, count);
                } else {
                    docCountMap.put(docName, old + count);
                }
            }

            StringBuilder sb = new StringBuilder();
            sb.append(total);
            for (Map.Entry<String, Integer> entry : docCountMap.entrySet()) {
                sb.append('\t').append(entry.getKey()).append(':').append(entry.getValue());
            }

            outValue.set(sb.toString());
            context.write(key, outValue);
        }
    }

    /** 第三阶段排序键：按总词频降序、词语升序 */
    public static class SortKey implements WritableComparable<SortKey> {
        private IntWritable total = new IntWritable();
        private Text word = new Text();

        public SortKey() {
        }

        public SortKey(int total, String word) {
            this.total.set(total);
            this.word.set(word);
        }

        public int getTotal() {
            return total.get();
        }

        public String getWord() {
            return word.toString();
        }

        @Override
        public void write(java.io.DataOutput out) throws IOException {
            total.write(out);
            word.write(out);
        }

        @Override
        public void readFields(java.io.DataInput in) throws IOException {
            total.readFields(in);
            word.readFields(in);
        }

        @Override
        public int compareTo(SortKey other) {
            int cmp = Integer.compare(other.total.get(), this.total.get());
            if (cmp != 0) {
                return cmp;
            }
            return this.word.compareTo(other.word);
        }

        @Override
        public int hashCode() {
            return total.hashCode() * 163 + word.hashCode();
        }

        @Override
        public boolean equals(Object obj) {
            if (obj == this) {
                return true;
            }
            if (!(obj instanceof SortKey)) {
                return false;
            }
            SortKey other = (SortKey) obj;
            return total.equals(other.total) && word.equals(other.word);
        }

        @Override
        public String toString() {
            return word.toString() + "\t" + total.get();
        }
    }

    /** 第三阶段 Mapper：解析第二阶段输出并转成可排序的 key */
    public static class SortMapper extends Mapper<LongWritable, Text, SortKey, Text> {
        private final Text outValue = new Text();

        @Override
        protected void map(LongWritable key, Text value, Context context) throws IOException, InterruptedException {
            String line = value.toString();
            if (line == null || line.isEmpty()) {
                return;
            }

            int firstTab = line.indexOf('\t');
            if (firstTab < 0 || firstTab == line.length() - 1) {
                return;
            }

            String word = line.substring(0, firstTab);
            String remainder = line.substring(firstTab + 1);
            int secondTab = remainder.indexOf('\t');
            String totalStr = secondTab >= 0 ? remainder.substring(0, secondTab) : remainder;
            int total = Integer.parseInt(totalStr.trim());

            outValue.set(remainder);
            context.write(new SortKey(total, word), outValue);
        }
    }

    /** 第三阶段 Reducer：输出最终结果 */
    public static class SortReducer extends Reducer<SortKey, Text, Text, Text> {
        private final Text outKey = new Text();

        @Override
        protected void reduce(SortKey key, Iterable<Text> values, Context context)
                throws IOException, InterruptedException {
            outKey.set(key.getWord());
            for (Text value : values) {
                context.write(outKey, value);
            }
        }
    }

    /** 可选：明确告诉 Hadoop 仍按 SortKey.compareTo 排序 */
    public static class SortKeyComparator extends WritableComparator {
        protected SortKeyComparator() {
            super(SortKey.class, true);
        }

        @Override
        public int compare(WritableComparable a, WritableComparable b) {
            return ((SortKey) a).compareTo((SortKey) b);
        }
    }

    @Override
    public int run(String[] args) throws Exception {
        if (args.length != 2 && args.length != 3) {
            System.err.println("用法: hadoop jar inverted-index.jar InvertedIndex <input> <output> [stopwords]");
            return 1;
        }

        String inputPath = args[0];
        String outputPath = args[1];
        boolean useStopWords = args.length == 3;
        String stopWordPath = useStopWords ? args[2] : null;

        Path temp1 = new Path(outputPath + "_tmp_stage1");
        Path temp2 = new Path(outputPath + "_tmp_stage2");
        Path finalOutput = new Path(outputPath);

        Configuration conf = getConf();
        FileSystem fs = FileSystem.get(conf);
        deleteIfExists(fs, temp1);
        deleteIfExists(fs, temp2);
        deleteIfExists(fs, finalOutput);

        // 第一阶段：统计每个词在每个文档中的词频
        Configuration conf1 = new Configuration(conf);
        conf1.setBoolean("invertedindex.use.stopwords", useStopWords);
        Job job1 = Job.getInstance(conf1, useStopWords ? "InvertedIndex-Stage1-Task2" : "InvertedIndex-Stage1-Task1");
        job1.setJarByClass(InvertedIndex.class);

        job1.setMapperClass(WordDocMapper.class);
        job1.setCombinerClass(SumReducer.class);
        job1.setReducerClass(SumReducer.class);

        job1.setMapOutputKeyClass(Text.class);
        job1.setMapOutputValueClass(IntWritable.class);
        job1.setOutputKeyClass(Text.class);
        job1.setOutputValueClass(IntWritable.class);

        job1.setInputFormatClass(TextInputFormat.class);
        job1.setOutputFormatClass(TextOutputFormat.class);

        FileInputFormat.addInputPath(job1, new Path(inputPath));
        FileOutputFormat.setOutputPath(job1, temp1);

        if (useStopWords) {
            job1.addCacheFile(new URI(stopWordPath + "#hit_stopwords.txt"));
        }

        if (!job1.waitForCompletion(true)) {
            deleteIfExists(fs, temp1);
            deleteIfExists(fs, temp2);
            return 1;
        }

        // 第二阶段：生成倒排索引行（未排序）
        Job job2 = Job.getInstance(new Configuration(conf), useStopWords ? "InvertedIndex-Stage2-Task2" : "InvertedIndex-Stage2-Task1");
        job2.setJarByClass(InvertedIndex.class);

        job2.setMapperClass(PostingListMapper.class);
        job2.setReducerClass(PostingListReducer.class);

        job2.setMapOutputKeyClass(Text.class);
        job2.setMapOutputValueClass(Text.class);
        job2.setOutputKeyClass(Text.class);
        job2.setOutputValueClass(Text.class);

        job2.setInputFormatClass(TextInputFormat.class);
        job2.setOutputFormatClass(TextOutputFormat.class);

        FileInputFormat.addInputPath(job2, temp1);
        FileOutputFormat.setOutputPath(job2, temp2);

        if (!job2.waitForCompletion(true)) {
            deleteIfExists(fs, temp1);
            deleteIfExists(fs, temp2);
            return 1;
        }

        // 第三阶段：按照总词频降序排序并输出最终结果
        Job job3 = Job.getInstance(new Configuration(conf), useStopWords ? "InvertedIndex-Stage3-Task2" : "InvertedIndex-Stage3-Task1");
        job3.setJarByClass(InvertedIndex.class);

        job3.setMapperClass(SortMapper.class);
        job3.setReducerClass(SortReducer.class);
        job3.setSortComparatorClass(SortKeyComparator.class);
        job3.setNumReduceTasks(1); // 保证最终输出全局有序

        job3.setMapOutputKeyClass(SortKey.class);
        job3.setMapOutputValueClass(Text.class);
        job3.setOutputKeyClass(Text.class);
        job3.setOutputValueClass(Text.class);

        job3.setInputFormatClass(TextInputFormat.class);
        job3.setOutputFormatClass(TextOutputFormat.class);

        FileInputFormat.addInputPath(job3, temp2);
        FileOutputFormat.setOutputPath(job3, finalOutput);

        boolean success = job3.waitForCompletion(true);

        // 清理临时目录
        deleteIfExists(fs, temp1);
        deleteIfExists(fs, temp2);

        return success ? 0 : 1;
    }

    private static void deleteIfExists(FileSystem fs, Path path) throws IOException {
        if (fs.exists(path)) {
            fs.delete(path, true);
        }
    }

    private static String stripExtension(String fileName) {
        int dot = fileName.lastIndexOf('.');
        if (dot > 0) {
            return fileName.substring(0, dot);
        }
        return fileName;
    }

    public static void main(String[] args) throws Exception {
        int exitCode = ToolRunner.run(new Configuration(), new InvertedIndex(), args);
        System.exit(exitCode);
    }
}
