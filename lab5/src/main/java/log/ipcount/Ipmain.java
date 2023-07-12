package log.ipcount;
import org.apache.hadoop.conf.Configured;
import org.apache.hadoop.fs.Path;
import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Job;
import org.apache.hadoop.mapreduce.lib.input.FileInputFormat;
import org.apache.hadoop.mapreduce.lib.output.FileOutputFormat;
import org.apache.hadoop.util.Tool;
import org.apache.hadoop.util.ToolRunner;

public class Ipmain extends Configured implements Tool  {
    @Override
    public int run(String[] args) throws Exception {
       Job job = Job.getInstance(getConf(), "ip-count");
        job.setJarByClass(getClass());
        job.setMapperClass(Ipmap.class);
        job.setReducerClass(Ipreduce.class);
        job.setCombinerClass(Ipcombiner.class);//combiner
        job.setMapOutputKeyClass(org.apache.hadoop.io.Text.class);
        job.setOutputKeyClass(org.apache.hadoop.io.Text.class);
        job.setMapOutputValueClass(Text.class);
        job.setOutputValueClass(LongWritable.class);
        FileInputFormat.setInputPaths(job, new Path(args[0]));
        FileOutputFormat.setOutputPath(job, new Path(args[1]));
        return job.waitForCompletion(true) ? 0 : 1;
    }
     public static void main(String[] args) throws Exception {
        int exitCode = ToolRunner.run(new Ipmain(), args);
        System.exit(exitCode);
    }
}
