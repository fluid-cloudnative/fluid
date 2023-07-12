package log.parse;


import java.io.IOException;

import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.NullWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Reducer;

public class ParseReducer extends Reducer<LongWritable, Text, NullWritable, Text> {
    
    @Override
    protected void reduce(LongWritable key, Iterable<Text> values,
            Reducer<LongWritable, Text, NullWritable, Text>.Context context) throws IOException, InterruptedException {
        for (Text record : values){
            context.write(NullWritable.get(), record);
        }
    }

}
