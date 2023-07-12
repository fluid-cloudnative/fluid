package log.ipal;

import java.io.IOException;
import log.ipal.IpUtil;
import org.apache.hadoop.io.LongWritable;
import org.apache.hadoop.io.Text;
import org.apache.hadoop.mapreduce.Mapper;

import java.io.IOException;

public class Ipalmap extends Mapper<LongWritable, Text,Text,LongWritable> {
    @Override
    protected void map(LongWritable key, Text value, Mapper<LongWritable, Text, Text, LongWritable>.Context context) throws IOException, InterruptedException {
       IpUtil temp=new IpUtil();//调用IpUtil方法
       String  splits[]=value.toString().split("\t");
       String address=temp.getCityInfo(splits[0]);//国家|区域|省份|城市|ISP

        //中国地区只输出省
        String add []=address.split("\\|");
        if("中国".equals(add[0].trim()))
        context.write(new Text(add[2]),new LongWritable(1));
        else context.write(new Text(add[0]),new LongWritable(1));
    }
}
