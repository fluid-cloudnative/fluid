package log.perhourvisit;

import org.apache.hadoop.conf.Configured;
import org.apache.hadoop.fs.Path;
import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Job;
import org.apache.hadoop.mapreduce.lib.input.FileInputFormat;
import org.apache.hadoop.mapreduce.lib.output.FileOutputFormat;
import org.apache.hadoop.util.Tool;
import org.apache.hadoop.util.ToolRunner;

public class Hourmain extends Configured implements Tool {
    @Override
    public int run(String[] args) throws Exception {
        Job job= Job.getInstance(getConf(),"perhour-visit");
        job.setJarByClass(getClass());
        job.setMapperClass(Hourmap.class);
        job.setReducerClass(Hourreduce.class);
        job.setCombinerClass(Hourreduce.class);//这里设置了combiner

        job.setMapOutputKeyClass(Text.class);
        job.setMapOutputValueClass(LongWritable.class);
        job.setOutputKeyClass(Text.class);
        job.setOutputValueClass(LongWritable.class);
        FileInputFormat.setInputPaths(job, new Path(args[0]));
        FileOutputFormat.setOutputPath(job, new Path(args[1]));
        return job.waitForCompletion(true) ? 0 : 1;

    }

    public static void main(String[] args) throws Exception {
        int exitcode= ToolRunner.run(new Hourmain(),args);
        System.exit(exitcode);
    }
}
