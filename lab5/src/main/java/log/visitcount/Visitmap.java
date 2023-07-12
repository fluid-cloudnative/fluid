package log.visitcount;
import java.io.IOException;

import org.apache.commons.daemon.support.DaemonLoader.Context;
import org.apache.hadoop.io.IntWritable;
import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Mapper;
public class Visitmap extends Mapper<LongWritable,Text,Text,LongWritable>{
    
    @Override
    protected void map(LongWritable key, Text value, Mapper<LongWritable, Text, Text, LongWritable>.Context context)
            throws IOException, InterruptedException {
       String splits[]=value.toString().split("\t");
        context.write(new Text(splits[5]),new LongWritable(1));
    }
}