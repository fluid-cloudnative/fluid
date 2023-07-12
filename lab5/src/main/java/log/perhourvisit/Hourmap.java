package log.perhourvisit;



import java.io.IOException;
import java.util.regex.Pattern;
import java.util.regex.Matcher;

import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Mapper;

public class Hourmap extends Mapper <LongWritable,Text,Text,LongWritable>{
    @Override
    protected void map(LongWritable key, Text value, Mapper<LongWritable, Text, Text, LongWritable>.Context context) throws IOException, InterruptedException {
        String splits[]=value.toString().split("\t");
        context.write(new Text(splits[4]),new LongWritable(1) );
    }
}
