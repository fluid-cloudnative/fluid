package log.ipal;

import log.perhourvisit.Hourmain;
import org.apache.hadoop.conf.Configured;
import org.apache.hadoop.fs.Path;
import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Job;
import org.apache.hadoop.mapreduce.lib.input.FileInputFormat;
import org.apache.hadoop.mapreduce.lib.output.FileOutputFormat;
import org.apache.hadoop.util.Tool;
import org.apache.hadoop.util.ToolRunner;

public class Ipalmain extends Configured implements Tool {
    @Override
    public int run(String[] args) throws Exception {
        Job job=Job.getInstance(getConf(),"ip-address");
        job.setJarByClass(getClass());
        job.setMapperClass(Ipalmap.class);
        job.setReducerClass(Ipalreduce.class);
        job.setCombinerClass(Ipalreduce.class);//combiner

        job.setMapOutputKeyClass(Text.class);
        job.setMapOutputValueClass(LongWritable.class);
        job.setOutputKeyClass(Text.class);
        job.setOutputValueClass(LongWritable.class);
        FileInputFormat.setInputPaths(job, new Path(args[0]));
        FileOutputFormat.setOutputPath(job, new Path(args[1]));
        return job.waitForCompletion(true) ? 0 : 1;
    }
    //test
    public static void main(String[] args) throws Exception {
        int exitcode= ToolRunner.run(new Ipalmain(),args);
        System.exit(exitcode);
    }
}
