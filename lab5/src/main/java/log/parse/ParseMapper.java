package log.parse;


import java.io.IOException;
import java.util.regex.Pattern;
import java.util.regex.Matcher;

import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Mapper;

public class ParseMapper extends Mapper<LongWritable, Text, LongWritable, Text> {
    
    Pattern pattern;

    @Override
    protected void setup(Mapper<LongWritable, Text, LongWritable, Text>.Context context)
            throws IOException, InterruptedException {
        String regex = "([0-9.]+).+\\[(\\d+)/(\\w+)/(\\d+):(\\d+).+\\] \"\\w* ?(\\S*).*\" (\\d+) (\\d+) \"(.*)\" \"(.*)\"";
        pattern = Pattern.compile(regex);
    }


    @Override
    protected void map(LongWritable key, Text value, Mapper<LongWritable, Text, LongWritable, Text>.Context context)
            throws IOException, InterruptedException {
        StringBuilder builder = new StringBuilder();
        Matcher matcher = pattern.matcher(value.toString());
        if (!matcher.matches()){
            System.out.println("!!!!!NO MATCH!!!!!");
            System.out.println(value.toString());
            System.exit(1);
        }
        for (int i = 1; i <= matcher.groupCount(); ++i){
            builder.append(matcher.group(i)).append('\t');
        }
        context.write(key, new Text(builder.toString()));
    }

}
