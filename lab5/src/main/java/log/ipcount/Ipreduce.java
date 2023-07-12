package log.ipcount;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Reducer;

public class Ipreduce extends Reducer<Text, Text, Text, LongWritable> {
    @Override
    protected void reduce(Text key, Iterable<Text> values, Reducer<Text, Text, Text, LongWritable>.Context context)
            throws IOException, InterruptedException {
        Map<String, Integer> map = new HashMap<>();// 用哈希表统计次数
        for (Text value : values) {
            String[] str = value.toString().split(",");// 用,分割
            for (int i = 0; i < str.length; i++) {
                if (map.containsKey(str[i])) {

                } else {
                    map.put(str[i], 1);// 不存在就插入
                }
            }
        }

        context.write(key, new LongWritable(map.size()));

    }
}
