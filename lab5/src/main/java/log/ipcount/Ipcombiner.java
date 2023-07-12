package log.ipcount;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.stream.Collectors;

import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Reducer;

public class Ipcombiner extends Reducer<Text, Text, Text, Text> {
    @Override
    protected void reduce(Text key, Iterable<Text> values, Reducer<Text, Text, Text, Text>.Context context)
            throws IOException, InterruptedException {
        Map<String, Integer> arr = new HashMap<>();// 用哈希表统计次数
        for (Text value : values) {
            String str = value.toString();
            if (arr.containsKey(str)) {

            } else {
                arr.put(str, 1);// 不存在就插入
            }
        }

        String keys = arr.keySet().stream()
                .collect(Collectors.joining(","));

        context.write(key, new Text(keys));
    }

}
